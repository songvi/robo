package store

import (
	"context"

	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/songvi/robo/models"
)

// Store defines the CRUD interface for all models
type Store interface {
	CreateJob(ctx context.Context, job *models.Job) error
	GetJob(ctx context.Context, id string) (*models.Job, error)
	UpdateJob(ctx context.Context, job *models.Job) error
	DeleteJob(ctx context.Context, id string) error
	GetJobsByStatus(ctx context.Context, status string, jobs *[]models.Job) error

	CreateWorker(ctx context.Context, worker *models.Worker) error
	GetWorker(ctx context.Context, id string) (*models.Worker, error)
	UpdateWorker(ctx context.Context, worker *models.Worker) error
	DeleteWorker(ctx context.Context, id string) error

	CreateUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, id string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id string) error

	CreateFile(ctx context.Context, file *models.File) error
	GetFile(ctx context.Context, id string) (*models.File, error)
	UpdateFile(ctx context.Context, file *models.File) error
	DeleteFile(ctx context.Context, id string) error

	CreateWorkspace(ctx context.Context, workspace *models.Workspace) error
	GetWorkspace(ctx context.Context, id string) (*models.Workspace, error)
	UpdateWorkspace(ctx context.Context, workspace *models.Workspace) error
	DeleteWorkspace(ctx context.Context, id string) error

	CreateCycle(ctx context.Context, cycle *models.Cycle) error
	GetCycle(ctx context.Context, id string) (*models.Cycle, error)
	UpdateCycle(ctx context.Context, cycle *models.Cycle) error
	DeleteCycle(ctx context.Context, id string) error
}

// GORMStore is the implementation of Store using GORM
type GORMStore struct {
	db *gorm.DB
}

// NewGORMStore initializes a new GORMStore
func NewGORMStore(db *gorm.DB) *GORMStore {
	return &GORMStore{db: db}
}

// CRUD methods for Job
func (s *GORMStore) CreateJob(ctx context.Context, job *models.Job) error {
	return s.db.WithContext(ctx).Create(job).Error
}

func (s *GORMStore) GetJob(ctx context.Context, id string) (*models.Job, error) {
	var job models.Job
	if err := s.db.WithContext(ctx).First(&job, "uuid = ?", id).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (s *GORMStore) UpdateJob(ctx context.Context, job *models.Job) error {
	return s.db.WithContext(ctx).Save(job).Error
}

func (s *GORMStore) DeleteJob(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&models.Job{}, "uuid = ?", id).Error
}

func (s *GORMStore) GetJobsByStatus(ctx context.Context, status string, jobs *[]models.Job) error {
	return s.db.WithContext(ctx).Where("status = ?", status).Find(jobs).Error
}

// CRUD methods for Worker
func (s *GORMStore) CreateWorker(ctx context.Context, worker *models.Worker) error {
	return s.db.WithContext(ctx).Create(worker).Error
}

func (s *GORMStore) GetWorker(ctx context.Context, id string) (*models.Worker, error) {
	var worker models.Worker
	if err := s.db.WithContext(ctx).First(&worker, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &worker, nil
}

func (s *GORMStore) UpdateWorker(ctx context.Context, worker *models.Worker) error {
	return s.db.WithContext(ctx).Save(worker).Error
}

func (s *GORMStore) DeleteWorker(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&models.Worker{}, "id = ?", id).Error
}

// CRUD methods for User
func (s *GORMStore) CreateUser(ctx context.Context, user *models.User) error {
	return s.db.WithContext(ctx).Create(user).Error
}

func (s *GORMStore) GetUser(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *GORMStore) UpdateUser(ctx context.Context, user *models.User) error {
	return s.db.WithContext(ctx).Save(user).Error
}

func (s *GORMStore) DeleteUser(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&models.User{}, "id = ?", id).Error
}

// CRUD methods for File
func (s *GORMStore) CreateFile(ctx context.Context, file *models.File) error {
	return s.db.WithContext(ctx).Create(file).Error
}

func (s *GORMStore) GetFile(ctx context.Context, id string) (*models.File, error) {
	var file models.File
	if err := s.db.WithContext(ctx).First(&file, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (s *GORMStore) UpdateFile(ctx context.Context, file *models.File) error {
	return s.db.WithContext(ctx).Save(file).Error
}

func (s *GORMStore) DeleteFile(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&models.File{}, "id = ?", id).Error
}

// CRUD methods for Workspace
func (s *GORMStore) CreateWorkspace(ctx context.Context, workspace *models.Workspace) error {
	return s.db.WithContext(ctx).Create(workspace).Error
}

func (s *GORMStore) GetWorkspace(ctx context.Context, id string) (*models.Workspace, error) {
	var workspace models.Workspace
	if err := s.db.WithContext(ctx).First(&workspace, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &workspace, nil
}

func (s *GORMStore) UpdateWorkspace(ctx context.Context, workspace *models.Workspace) error {
	return s.db.WithContext(ctx).Save(workspace).Error
}

func (s *GORMStore) DeleteWorkspace(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&models.Workspace{}, "id = ?", id).Error
}

// CRUD methods for Cycle
func (s *GORMStore) CreateCycle(ctx context.Context, cycle *models.Cycle) error {
	return s.db.WithContext(ctx).Create(cycle).Error
}

func (s *GORMStore) GetCycle(ctx context.Context, id string) (*models.Cycle, error) {
	var cycle models.Cycle
	if err := s.db.WithContext(ctx).First(&cycle, "uuid = ?", id).Error; err != nil {
		return nil, err
	}
	return &cycle, nil
}

func (s *GORMStore) UpdateCycle(ctx context.Context, cycle *models.Cycle) error {
	return s.db.WithContext(ctx).Save(cycle).Error
}

func (s *GORMStore) DeleteCycle(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&models.Cycle{}, "uuid = ?", id).Error
}

// ProvideStore is an fx-compatible constructor
func ProvideStore(lc fx.Lifecycle, db *gorm.DB) Store {
	store := NewGORMStore(db)

	// Add lifecycle hooks for migrations
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Perform migrations for database tables
			return db.AutoMigrate(
				&models.Job{},
				&models.Worker{},
				&models.User{},
				&models.File{},
				&models.Workspace{},
				&models.Cycle{},
			)
		},
		OnStop: func(ctx context.Context) error {
			// Cleanup tasks if needed
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			return sqlDB.Close()
		},
	})

	return store
}

// Module exports the Store for fx
var Module = fx.Provide(ProvideStore)
