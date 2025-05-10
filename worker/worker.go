package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/fx"

	"github.com/songvi/robo/config"
	"github.com/songvi/robo/logger"
)

// Job defines the structure of a job (same as dispatcher)
type Job struct {
	UUID       string          `json:"uuid" yaml:"uuid"`
	WorkerID   string          `json:"worker_id" yaml:"worker_id"`
	Name       string          `json:"name" yaml:"name"`
	InputData  json.RawMessage `json:"input_data" yaml:"input_data"`
	OutputData json.RawMessage `json:"output_data" yaml:"output_data"`
	Error      string          `json:"error" yaml:"error"`
	StartAt    int64           `json:"start_at" yaml:"start_at"`
	DoneAt     int64           `json:"done_at" yaml:"done_at"`
	Status     string          `json:"status" yaml:"status"`
}

// Worker defines the worker service
type Worker interface {
	Start(ctx context.Context) error
}

// workerImpl implements the Worker interface
type workerImpl struct {
	nc       *nats.Conn
	logger   logger.Logger
	config   config.ConfigService
	workerID string
	name     string
}

// NewWorker creates a new Worker instance
func NewWorker(lc fx.Lifecycle, config config.ConfigService, logger logger.Logger, nc *nats.Conn) Worker {
	w := &workerImpl{
		nc:       nc,
		logger:   logger,
		config:   config,
		workerID: "worker-1", // Should be unique, e.g., generated UUID
		name:     "Worker1",
	}

	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			logger.Debug(ctx, "Starting worker", "worker_id", w.workerID)
			return w.Start(ctx)
		},
		OnStop: func(context.Context) error {
			logger.Debug(ctx, "Stopping worker", "worker_id", w.workerID)
			cancel()
			return nil
		},
	})

	return w
}

// Start begins worker operations
func (w *workerImpl) Start(ctx context.Context) error {
	// Register worker
	regMsg := struct {
		WorkerID     string   `json:"worker_id"`
		Name         string   `json:"name"`
		Capabilities []string `json:"capabilities"`
		Status       string   `json:"status"`
	}{
		WorkerID:     w.workerID,
		Name:         w.name,
		Capabilities: []string{"file_processing", "task_execution"},
		Status:       "registered",
	}
	data, err := json.Marshal(regMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal registration message: %w", err)
	}
	if err := w.nc.Publish("dispatcher.worker.register", data); err != nil {
		return fmt.Errorf("failed to publish registration: %w", err)
	}
	w.logger.Info(ctx, "Worker registered", "worker_id", w.workerID, "name", w.name)

	// Subscribe to jobs
	jobSubject := fmt.Sprintf("dispatcher.job.%s", w.workerID)
	jobCh, err := w.subscribe(ctx, jobSubject)
	if err != nil {
		return fmt.Errorf("failed to subscribe to jobs: %w", err)
	}
	go w.handleJobs(ctx, jobCh)

	// Start heartbeat
	go w.sendHeartbeats(ctx)

	return nil
}

// subscribe subscribes to a NATS subject
func (w *workerImpl) subscribe(ctx context.Context, subject string) (<-chan *nats.Msg, error) {
	msgCh := make(chan *nats.Msg, 64)
	sub, err := w.nc.ChanSubscribe(subject, msgCh)
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		if err := sub.Unsubscribe(); err != nil {
			w.logger.Error(ctx, "Failed to unsubscribe from subject", "subject", subject, "error", err)
		}
		close(msgCh)
	}()
	w.logger.Info(ctx, "Subscribed to subject", "subject", subject)
	return msgCh, nil
}

// handleJobs processes incoming jobs
func (w *workerImpl) handleJobs(ctx context.Context, jobCh <-chan *nats.Msg) {
	for msg := range jobCh {
		var job Job
		if err := json.Unmarshal(msg.Data, &job); err != nil {
			w.logger.Error(ctx, "Failed to unmarshal job", "error", err)
			continue
		}
		w.logger.Info(ctx, "Received job", "job_uuid", job.UUID, "job_name", job.Name)

		// Process the job (placeholder logic)
		job.StartAt = time.Now().Unix()
		job.Status = "processing"
		// Example: Process InputData and set OutputData
		job.OutputData = []byte(`{"result":"processed"}`)
		job.Status = "completed"
		job.DoneAt = time.Now().Unix()

		// Publish result
		resultData, err := json.Marshal(job)
		if err != nil {
			w.logger.Error(ctx, "Failed to marshal job result", "job_uuid", job.UUID, "error", err)
			continue
		}
		if err := w.nc.Publish("dispatcher.job.result", resultData); err != nil {
			w.logger.Error(ctx, "Failed to publish job result", "job_uuid", job.UUID, "error", err)
			continue
		}
		w.logger.Info(ctx, "Job completed", "job_uuid", job.UUID, "worker_id", job.WorkerID)
	}
}

// sendHeartbeats sends periodic heartbeats
func (w *workerImpl) sendHeartbeats(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hbMsg := struct {
				WorkerID string `json:"worker_id"`
				Status   string `json:"status"`
			}{
				WorkerID: w.workerID,
				Status:   "heartbeat",
			}
			data, err := json.Marshal(hbMsg)
			if err != nil {
				w.logger.Error(ctx, "Failed to marshal heartbeat", "error", err)
				continue
			}
			if err := w.nc.Publish("dispatcher.worker.heartbeat", data); err != nil {
				w.logger.Error(ctx, "Failed to publish heartbeat", "error", err)
				continue
			}
			w.logger.Debug(ctx, "Sent heartbeat", "worker_id", w.workerID)
		}
	}
}
