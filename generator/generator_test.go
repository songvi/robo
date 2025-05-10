package generator

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/songvi/robo/models"
)

func TestGenerator(t *testing.T) {
	// Setup test database
	db, err := gorm.Open(sqlite.Open("file:.test/test.db?cache=shared&mode=rwc"), &gorm.Config{})
	require.NoError(t, err, "failed to open test database")

	// Create users table and insert test UUIDs
	err = db.AutoMigrate(&models.User{})
	require.NoError(t, err, "failed to migrate user table")
	testUUIDs := []string{
		"550e8400-e29b-41d4-a716-446655440000",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"7d793037-a076-4e26-b8e6-2d6f0cb3b3a7",
		"8f8b8f8b-8f8b-8f8b-8f8b-8f8b8f8b8f8b",
		"9e107d9d-372b-4b1b-a8f7-0c7e2f0b1c2d",
	}
	for _, uuid := range testUUIDs {
		db.Create(&models.User{UUID: uuid})
	}

	// Cleanup database after test
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
		os.Remove(".test/test.db")
	}()

	// Setup test file directory
	fileDir := ".test/files"
	err = os.MkdirAll(fileDir, 0755)
	require.NoError(t, err, "failed to create test file directory")
	defer os.RemoveAll(fileDir)

	// Define test config
	config := GeneratorConfig{
		Strategy: Strategy{
			UserStrategy: models.UserStrategy{
				UserLang:        []string{"en", "fr"},
				LangProbability: []float64{0.5, 0.3},
			},
			FileStrategy: models.FileStrategy{
				FileExtension:            []string{"txt", "jpeg", "bin"},
				FileExtensionProbability: []float64{0.1, 0.3, 0.6},
				FileSize:                 []int{1024, 1048576},
				FileSizeProbability:      []float64{0.7, 0.3},
				FileLang:                 []string{"en", "fr"},
				FileLangNameProbability:  []float64{0.5, 0.3},
			},
			WorkspaceStrategy: models.WorkspaceStrategy{
				NumberOfUsers:            []int{2, 3},
				NumberOfUsersProbability: []float64{0.6, 0.4},
			},
		},
		FileStore: FileStore{
			FilePath: fileDir,
		},
		DBStore: DBStore{},
		DBConfig: DBConfig{
			DSN: "file:.test/test.db?cache=shared&mode=rwc",
		},
		FileBuffer:      2,
		UserBuffer:      3,
		WorkspaceBuffer: 2,
	}

	// Create Fx app for testing
	var generator Generator
	app := fx.New(
		fx.Provide(func() GeneratorConfig { return config }),
		Module,
		fx.Populate(&generator),
	)

	// Start the app
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = app.Start(ctx)
	require.NoError(t, err, "failed to start Fx app")
	defer app.Stop(context.Background())

	t.Run("TestUsers", func(t *testing.T) {
		userCh := generator.Users(ctx)
		usersReceived := 0
		timeout := time.After(2 * time.Second)

		for usersReceived < config.UserBuffer {
			select {
			case u, ok := <-userCh:
				if !ok {
					t.Fatal("user channel closed unexpectedly")
				}
				assert.NotEmpty(t, u.DisplayName, "user display name should not be empty")
				assert.NotEmpty(t, u.UserName, "user username should not be empty")
				assert.Contains(t, []string{"en", "fr"}, u.Language, "user language should be valid")
				usersReceived++
			case <-timeout:
				t.Fatalf("timed out waiting for users; received %d/%d", usersReceived, config.UserBuffer)
			}
		}
	})

	t.Run("TestFiles", func(t *testing.T) {
		fileCh := generator.Files(ctx)
		filesReceived := 0
		timeout := time.After(2 * time.Second)

		for filesReceived < config.FileBuffer {
			select {
			case f, ok := <-fileCh:
				if !ok {
					t.Fatal("file channel closed unexpectedly")
				}
				log.Printf("Received file info: %+v\n", f)
				assert.NotEmpty(t, f.Name, "file name should not be empty")
				assert.Contains(t, []string{"txt", "jpeg", "bin"}, f.FileExtension, "file extension should be valid")
				assert.Contains(t, []int{1024, 1048576}, f.FileSize, "file size should be valid")
				assert.NotEmpty(t, f.FileContent, "file content should point to a physical file")
				// Check if file exists
				filePath := filepath.Join(config.FileStore.FilePath, f.Name+"."+f.FileExtension)
				_, err := os.Stat(filePath)
				assert.NoError(t, err, "physical file should exist")
				filesReceived++
			case <-timeout:
				t.Fatalf("timed out waiting for files; received %d/%d", filesReceived, config.FileBuffer)
			}
		}
	})

	t.Run("TestWorkspaces", func(t *testing.T) {
		workspaceCh := generator.Workspaces(ctx)
		workspacesReceived := 0
		timeout := time.After(2 * time.Second)

		for workspacesReceived < config.WorkspaceBuffer {
			select {
			case w, ok := <-workspaceCh:
				if !ok {
					t.Fatal("workspace channel closed unexpectedly")
				}
				assert.NotEmpty(t, w.Name, "workspace name should not be empty")
				assert.Contains(t, []int{2, 3}, len(w.Users), "workspace user count should be valid")
				for _, uuid := range w.Users {
					assert.Contains(t, testUUIDs, uuid, "workspace user UUID should be from database")
				}
				workspacesReceived++
			case <-timeout:
				t.Fatalf("timed out waiting for workspaces; received %d/%d", workspacesReceived, config.WorkspaceBuffer)
			}
		}
	})
}
