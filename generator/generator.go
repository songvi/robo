package generator

import (
	"context"
	"log"
	"sync"

	"go.uber.org/fx"
	"gorm.io/driver/sqlite" // Example driver; replace with your database driver
	"gorm.io/gorm"

	"github.com/songvi/robo/generator/file"
	"github.com/songvi/robo/generator/user"
	"github.com/songvi/robo/generator/workspace"
)

// Generator defines the interface for the generator service
type Generator interface {
	Users(ctx context.Context) <-chan user.User
	Files(ctx context.Context) <-chan file.File
	Workspaces(ctx context.Context) <-chan workspace.Workspace
}

// generatorImpl is the implementation of the Generator interface
type generatorImpl struct {
	config        GeneratorConfig
	db            *gorm.DB
	userCh        chan user.User
	fileCh        chan file.File
	workspaceCh   chan workspace.Workspace
	wg            sync.WaitGroup
	cancelWorkers context.CancelFunc
}

// UserDBModel represents the user table in the database
type UserDBModel struct {
	UUID string `gorm:"primaryKey"`
}

// NewGenerator creates a new Generator instance with the provided config
func NewGenerator(lc fx.Lifecycle, config GeneratorConfig) (Generator, error) {
	// Initialize GORM database
	db, err := gorm.Open(sqlite.Open(config.DBConfig.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Set default buffer sizes if not specified
	userBuffer := config.UserBuffer
	if userBuffer <= 0 {
		userBuffer = 10 // Default buffer for users
	}
	fileBuffer := config.FileBuffer
	if fileBuffer <= 0 {
		fileBuffer = 5 // Default buffer for files (smaller due to potential large size)
	}
	workspaceBuffer := config.WorkspaceBuffer
	if workspaceBuffer <= 0 {
		workspaceBuffer = 10 // Default buffer for workspaces
	}

	g := &generatorImpl{
		config:      config,
		db:          db,
		userCh:      make(chan user.User, userBuffer),
		fileCh:      make(chan file.File, fileBuffer),
		workspaceCh: make(chan workspace.Workspace, workspaceBuffer),
	}

	// Create a context for worker cancellation
	ctx, cancel := context.WithCancel(context.Background())
	g.cancelWorkers = cancel

	// Start workers on Fx lifecycle start
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			g.startWorkers(ctx)
			return nil
		},
		OnStop: func(context.Context) error {
			g.stopWorkers()
			return nil
		},
	})

	return g, nil
}

// startWorkers starts the background workers for generating users, files, and workspaces
func (g *generatorImpl) startWorkers(ctx context.Context) {
	// User worker
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				user, err := GenerateUser(g.config.Strategy.UserStrategy)
				if err != nil {
					log.Printf("Error generating user: %v", err)
					continue // Log error in production
				}
				select {
				case g.userCh <- user:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	// File worker
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				file, err := GenerateFile(g.config.Strategy.FileStrategy, g.config.FileStore.FilePath)
				if err != nil {
					continue // Log error in production
				}
				select {
				case g.fileCh <- file:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	// Workspace worker
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// Fetch UUIDs from database
				var users []UserDBModel
				// Get the maximum number of users needed based on WorkspaceStrategy
				maxUsers := max(g.config.Strategy.WorkspaceStrategy.NumberOfUsers)
				if err := g.db.Limit(maxUsers).Find(&users).Error; err != nil {
					continue // Log error in production
				}
				if len(users) == 0 {
					continue // No users available; retry
				}

				// Extract UUIDs
				uuids := make([]string, len(users))
				for i, u := range users {
					uuids[i] = u.UUID
				}

				// Generate workspace
				workspace, err := GenerateWorkspace(g.config.Strategy.WorkspaceStrategy, uuids)
				if err != nil {
					continue // Log error in production
				}
				select {
				case g.workspaceCh <- workspace:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
}

// max returns the maximum value in a slice of integers
func max(numbers []int) int {
	if len(numbers) == 0 {
		return 0
	}
	maxVal := numbers[0]
	for _, n := range numbers[1:] {
		if n > maxVal {
			maxVal = n
		}
	}
	return maxVal
}

// stopWorkers stops all background workers and closes channels
func (g *generatorImpl) stopWorkers() {
	g.cancelWorkers()
	g.wg.Wait()
	close(g.userCh)
	close(g.fileCh)
	close(g.workspaceCh)
	// Close database connection
	sqlDB, _ := g.db.DB()
	sqlDB.Close()
}

// Users returns a channel of generated users
func (g *generatorImpl) Users(ctx context.Context) <-chan user.User {
	return g.userCh
}

// Files returns a channel of generated files
func (g *generatorImpl) Files(ctx context.Context) <-chan file.File {
	return g.fileCh
}

// Workspaces returns a channel of generated workspaces
func (g *generatorImpl) Workspaces(ctx context.Context) <-chan workspace.Workspace {
	return g.workspaceCh
}

// Module defines the Fx module for the Generator service
var Module = fx.Module(
	"generator",
	fx.Provide(
		NewGenerator,
	),
)
