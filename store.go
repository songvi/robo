package main

import (
	"context"

	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/songvi/robo/dispatcher"
	"github.com/songvi/robo/generator/file"
	"github.com/songvi/robo/generator/user"
	"github.com/songvi/robo/generator/workspace"
	"github.com/songvi/robo/job"
)

// Store defines the CRUD interface for all models
type Store interface {
	CreateJob(ctx context.Context, job *job.Job) error
	GetJob(ctx context.Context, id string) (*job.Job, error)
	UpdateJob(ctx context.Context, job *job.Job) error
	DeleteJob(ctx context.Context, id string) error

	CreateWorker(ctx context.Context, worker *dispatcher.Worker) error
	GetWorker(ctx context.Context, id string) (*dispatcher.Worker, error)
	UpdateWorker(ctx context.Context, worker *dispatcher.Worker) error
	DeleteWorker(ctx context.Context, id string) error

	CreateUser(ctx context.Context, user *user.User) error
	GetUser(ctx context.Context, id string) (*user.User, error)
	UpdateUser(ctx context.Context, user *user.User) error
	DeleteUser(ctx context.Context, id string) error

	CreateFile(ctx context.Context, file *file.File) error
	GetFile(ctx context.Context, id string) (*file.File, error)
	UpdateFile(ctx context.Context, file *file.File) error
	DeleteFile(ctx context.Context, id string) error

	CreateWorkspace(ctx context.Context, workspace *workspace.Workspace) error
	GetWorkspace(ctx context.Context, id string) (*workspace.Workspace, error)
	UpdateWorkspace(ctx context.Context, workspace *workspace.Workspace) error
	DeleteWorkspace(ctx context.Context, id string) error
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
func (s *GORMStore) CreateJob(ctx context.Context, job *job.Job) error {
	return s.db.WithContext(ctx).Create(job).Error
}

func (s *GORMStore) GetJob(ctx context.Context, id string) (*job.Job, error) {
	var job job.Job
	if err := s.db.WithContext(ctx).First(&job, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (s *GORMStore) UpdateJob(ctx context.Context, job *job.Job) error {
	return s.db.WithContext(ctx).Save(job).Error
}

func (s *GORMStore) DeleteJob(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&job.Job{}, "id = ?", id).Error
}

// Similar CRUD methods for Worker, User, File, and Workspace ...

// CRUD methods for Worker
func (s *GORMStore) CreateWorker(ctx context.Context, worker *dispatcher.Worker) error {
	return s.db.WithContext(ctx).Create(worker).Error
}

func (s *GORMStore) GetWorker(ctx context.Context, id string) (*dispatcher.Worker, error) {
	var worker dispatcher.Worker
	if err := s.db.WithContext(ctx).First(&worker, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &worker, nil
}

func (s *GORMStore) UpdateWorker(ctx context.Context, worker *dispatcher.Worker) error {
	return s.db.WithContext(ctx).Save(worker).Error
}

func (s *GORMStore) DeleteWorker(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&dispatcher.Worker{}, "id = ?", id).Error
}

// CRUD methods for User
func (s *GORMStore) CreateUser(ctx context.Context, user *user.User) error {
	return s.db.WithContext(ctx).Create(user).Error
}

func (s *GORMStore) GetUser(ctx context.Context, id string) (*user.User, error) {
	var usr user.User
	if err := s.db.WithContext(ctx).First(&usr, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &usr, nil
}

func (s *GORMStore) UpdateUser(ctx context.Context, user *user.User) error {
	return s.db.WithContext(ctx).Save(user).Error
}

func (s *GORMStore) DeleteUser(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&user.User{}, "id = ?", id).Error
}

// CRUD methods for File
func (s *GORMStore) CreateFile(ctx context.Context, file *file.File) error {
	return s.db.WithContext(ctx).Create(file).Error
}

func (s *GORMStore) GetFile(ctx context.Context, id string) (*file.File, error) {
	var fl file.File
	if err := s.db.WithContext(ctx).First(&fl, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &fl, nil
}

func (s *GORMStore) UpdateFile(ctx context.Context, file *file.File) error {
	return s.db.WithContext(ctx).Save(file).Error
}

func (s *GORMStore) DeleteFile(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&file.File{}, "id = ?", id).Error
}

// CRUD methods for Workspace
func (s *GORMStore) CreateWorkspace(ctx context.Context, workspace *workspace.Workspace) error {
	return s.db.WithContext(ctx).Create(workspace).Error
}

func (s *GORMStore) GetWorkspace(ctx context.Context, id string) (*workspace.Workspace, error) {
	var ws workspace.Workspace
	if err := s.db.WithContext(ctx).First(&ws, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &ws, nil
}

func (s *GORMStore) UpdateWorkspace(ctx context.Context, workspace *workspace.Workspace) error {
	return s.db.WithContext(ctx).Save(workspace).Error
}

func (s *GORMStore) DeleteWorkspace(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&workspace.Workspace{}, "id = ?", id).Error
}

// ProvideStore is an fx-compatible constructor
func ProvideStore(lc fx.Lifecycle, db *gorm.DB) Store {
	store := NewGORMStore(db)

	// Add lifecycle hooks if needed
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Perform any startup tasks such as migrations
			return db.AutoMigrate(&job.Job{},
				&dispatcher.Worker{},
				&user.User{},
				&file.File{},
				&workspace.Workspace{},
				&dispatcher.Cycle{},
				&dispatcher.Strategy{},
			)
		},
		OnStop: func(ctx context.Context) error {
			// Cleanup tasks if needed
			return nil
		},
	})

	return store
}

// Module exports the Store for fx
var Module = fx.Provide(ProvideStore)
