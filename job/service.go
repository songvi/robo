package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"go.uber.org/fx"

	"github.com/songvi/robo/config"
	"github.com/songvi/robo/dispatcher"
	"github.com/songvi/robo/generator"
	"github.com/songvi/robo/logger"
	"github.com/songvi/robo/models"
	"github.com/songvi/robo/store"
)

// JobService defines the interface for job management
type JobService interface {
	StartCycle(ctx context.Context, cycle models.Cycle) error
	ProcessJobs(ctx context.Context) error
}

// jobServiceImpl implements the JobService interface
type jobServiceImpl struct {
	store      store.Store
	dispatcher dispatcher.Dispatcher
	logger     logger.Logger
	config     config.ConfigService
	generator  generator.Generator
}

// NewJobService creates a new JobService instance
func NewJobService(
	lc fx.Lifecycle,
	config config.ConfigService,
	logger logger.Logger,
	store store.Store,
	dispatcher dispatcher.Dispatcher,
	generator generator.Generator,
) JobService {
	s := &jobServiceImpl{
		store:      store,
		dispatcher: dispatcher,
		logger:     logger,
		config:     config,
		generator:  generator,
	}

	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			logger.Info(ctx, "Starting JobService")
			go s.ProcessJobs(ctx)
			return nil
		},
		OnStop: func(context.Context) error {
			logger.Info(ctx, "Stopping JobService")
			cancel()
			return nil
		},
	})

	return s
}

// StartCycle initiates a new cycle and generates sessions and jobs
func (s *jobServiceImpl) StartCycle(ctx context.Context, cycle models.Cycle) error {
	cycle.UUID = uuid.New().String()
	cycle.StartedAt = time.Now().Unix()
	cycle.Status = "running"

	// Save cycle to database
	if err := s.store.CreateCycle(ctx, &cycle); err != nil {
		s.logger.Error(ctx, "Failed to save cycle to database", "cycle_uuid", cycle.UUID, "error", err)
		return err
	}

	// Fetch users from generator
	users := []models.User{}
	userCh := s.generator.Users(ctx)
	for i := 0; i < cycle.Strategy.MaxUsers; i++ {
		select {
		case user, ok := <-userCh:
			if !ok {
				break
			}
			users = append(users, user)
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	for _, user := range users {
		session := models.Session{UserID: user.UserName}
		// Generate jobs for the session
		jobs, err := s.generateSessionJobs(ctx, cycle, session)
		if err != nil {
			s.logger.Error(ctx, "Failed to generate jobs for session", "cycle_uuid", cycle.UUID, "user_id", session.UserID, "error", err)
			continue
		}

		// Save jobs to database
		for _, job := range jobs {
			job.CycleUUID = cycle.UUID
			job.SessionID = session.UserID
			if err := s.store.CreateJob(ctx, &job); err != nil {
				s.logger.Error(ctx, "Failed to save job to database", "job_uuid", job.UUID, "error", err)
				continue
			}
		}
	}

	s.logger.Info(ctx, "Cycle started", "cycle_uuid", cycle.UUID, "name", cycle.Name)
	return nil
}

// generateSessionJobs creates jobs for a session
func (s *jobServiceImpl) generateSessionJobs(ctx context.Context, cycle models.Cycle, session models.Session) ([]models.Job, error) {
	var jobs []models.Job
	actions := []string{
		"create_user", "update_user", "delete_user",
		"create_workspace", "update_workspace", "delete_workspace",
		"upload_file", "download_file", "consult_file",
	}

	// Generate jobs based on strategy limits
	totalJobs := cycle.Strategy.MaxFiles + cycle.Strategy.MaxWorkspaces
	for i := 0; i < totalJobs; i++ {
		action := actions[i%len(actions)]
		inputData := map[string]string{
			"user_id": session.UserID,
			"action":  action,
		}
		inputJSON, err := json.Marshal(inputData)
		if err != nil {
			s.logger.Error(ctx, "Failed to marshal job input data", "action", action, "error", err)
			continue
		}

		job := models.Job{
			UUID:      uuid.New().String(),
			Name:      action,
			InputData: json.RawMessage(inputJSON),
			Status:    "pending",
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// ProcessJobs dispatches pending jobs and processes results
func (s *jobServiceImpl) ProcessJobs(ctx context.Context) error {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			// Fetch pending jobs
			var jobs []models.Job
			if err := s.store.GetJobsByStatus(ctx, "pending", &jobs); err != nil {
				s.logger.Error(ctx, "Failed to fetch pending jobs", "error", err)
				continue
			}

			for _, job := range jobs {
				// Dispatch job
				if err := s.dispatcher.DispatchJob(ctx, &job); err != nil {
					s.logger.Error(ctx, "Failed to dispatch job", "job_uuid", job.UUID, "error", err)
					continue
				}

				// Update job status
				job.Status = "dispatched"
				if err := s.store.UpdateJob(ctx, &job); err != nil {
					s.logger.Error(ctx, "Failed to update job status", "job_uuid", job.UUID, "error", err)
					continue
				}
			}

			// Process job results
			resultCh, err := s.dispatcher.Subscribe(ctx, "dispatcher.job.result")
			if err != nil {
				s.logger.Error(ctx, "Failed to subscribe to job results", "error", err)
				continue
			}
			go func() {
				for msg := range resultCh {
					var job models.Job
					if err := json.Unmarshal(msg.Data, &job); err != nil {
						s.logger.Error(ctx, "Failed to unmarshal job result", "error", err)
						continue
					}

					// Update job result in database
					if err := s.store.UpdateJob(ctx, &job); err != nil {
						s.logger.Error(ctx, "Failed to save job result", "job_uuid", job.UUID, "error", err)
						continue
					}

					s.logger.Info(ctx, "Job result processed", "job_uuid", job.UUID, "status", job.Status)

					// Check if cycle is complete
					if err := s.checkCycleCompletion(ctx, job.CycleUUID); err != nil {
						s.logger.Error(ctx, "Failed to check cycle completion", "cycle_uuid", job.CycleUUID, "error", err)
					}
				}
			}()
		}
	}
}

// // checkCycleCompletion checks if all jobs in a cycle are complete
// func (s *jobServiceImpl) ProcessJobs(ctx context.Context) error {
// 	ticker := time.NewTicker(10 * time.Second)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return nil
// 		case <-ticker.C:
// 			// Fetch pending jobs
// 			var jobs []models.Job
// 			if err := s.store.GetJobsByStatus(ctx, "pending", &jobs); err != nil {
// 				s.logger.Error(ctx, "Failed to fetch pending jobs", "error", err)
// 				continue
// 			}

// 			for _, job := range jobs {
// 				// Dispatch job
// 				if err := s.dispatcher.DispatchJob(ctx, &job); err != nil {
// 					s.logger.Error(ctx, "Failed to dispatch job", "job_uuid", job.UUID, "error", err)
// 					continue
// 				}

// 				// Update job status
// 				job.Status = "dispatched"
// 				if err := s.store.UpdateJob(ctx, &job); err != nil {
// 					s.logger.Error(ctx, "Failed to update job status", "job_uuid", job.UUID, "error", err)
// 					continue
// 				}
// 			}

// 			// Process job results
// 			resultCh, err := s.dispatcher.Subscribe(ctx, "dispatcher.job.result")
// 			if err != nil {
// 				s.logger.Error(ctx, "Failed to subscribe to job results", "error", err)
// 				continue
// 			}
// 			go func() {
// 				for msg := range resultCh {
// 					var job models.Job
// 					if err := json.Unmarshal(msg.Data, &job); err != nil {
// 						s.logger.Error(ctx, "Failed to unmarshal job result", "error", err)
// 						continue
// 					}

// 					// Update job result in database
// 					if err := s.store.UpdateJob(ctx, &job); err != nil {
// 						s.logger.Error(ctx, "Failed to save job result", "job_uuid", job.UUID, "error", err)
// 						continue
// 					}

// 					s.logger.Info(ctx, "Job result processed", "job_uuid", job.UUID, "status", job.Status)

// 					// Check if cycle is complete
// 					if err := s.checkCycleCompletion(ctx, job.CycleUUID); err != nil {
// 						s.logger.Error(ctx, "Failed to check cycle completion", "cycle_uuid", job.CycleUUID, "error", err)
// 					}
// 				}
// 			}()
// 		}
// 	}
// }

// checkCycleCompletion checks if all jobs in a cycle are complete
func (s *jobServiceImpl) checkCycleCompletion(ctx context.Context, cycleUUID string) error {
	var pendingJobs []models.Job
	if err := s.store.GetJobsByStatus(ctx, "pending", &pendingJobs); err != nil {
		return err
	}
	var dispatchedJobs []models.Job
	if err := s.store.GetJobsByStatus(ctx, "dispatched", &dispatchedJobs); err != nil {
		return err
	}

	if len(pendingJobs)+len(dispatchedJobs) == 0 {
		cycle, err := s.store.GetCycle(ctx, cycleUUID)
		if err != nil {
			return err
		}
		cycle.Status = "completed"
		cycle.DoneAt = time.Now().Unix()
		if err := s.store.UpdateCycle(ctx, cycle); err != nil {
			return err
		}
		s.logger.Info(ctx, "Cycle completed", "cycle_uuid", cycleUUID)
	}

	return nil
}

// Module defines the Fx module for the JobService
func Module(lc fx.Lifecycle, config config.ConfigService, logger logger.Logger, store store.Store, dispatcher dispatcher.Dispatcher, generator generator.Generator) fx.Option {
	return fx.Module(
		"service",
		fx.Provide(NewJobService),
		fx.Invoke(func(s JobService) {
			// Ensure JobService is instantiated
			logger.Info(context.Background(), "JobService module initialized")
		}),
	)
}
