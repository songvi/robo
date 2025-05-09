package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/fx"

	"github.com/songvi/robo/logger"
)

// Worker defines the worker service
type Worker struct {
	id           string
	name         string
	nc           *nats.Conn
	logger       logger.Logger
	ctx          context.Context
	cancel       context.CancelFunc
	capabilities []string
}

// NewWorker creates a new Worker instance
func NewWorker(lc fx.Lifecycle, id, name string, nc *nats.Conn, logger logger.Logger) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	w := &Worker{
		id:           id,
		name:         name,
		nc:           nc,
		logger:       logger,
		ctx:          ctx,
		cancel:       cancel,
		capabilities: []string{"file_processing", "task_execution"},
	}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			return w.Start()
		},
		OnStop: func(context.Context) error {
			return w.Stop()
		},
	})

	return w
}

// Register sends a registration message to the dispatcher
func (w *Worker) Register() error {
	if !w.nc.IsConnected() {
		w.logger.Error(w.ctx, "Cannot register: NATS not connected", "worker_id", w.id)
		return fmt.Errorf("cannot register worker %s: NATS not connected", w.id)
	}
	msg := struct {
		WorkerID     string   `json:"worker_id"`
		Name         string   `json:"name"`
		Capabilities []string `json:"capabilities"`
		Status       string   `json:"status"`
	}{
		WorkerID:     w.id,
		Name:         w.name,
		Capabilities: w.capabilities,
		Status:       "registered",
	}
	data, err := json.Marshal(msg)
	if err != nil {
		w.logger.Error(w.ctx, "Failed to marshal registration message", "worker_id", w.id, "error", err)
		return fmt.Errorf("failed to marshal registration message: %w", err)
	}
	if err := w.nc.Publish("dispatcher.worker.register", data); err != nil {
		w.logger.Error(w.ctx, "Failed to publish registration message", "worker_id", w.id, "error", err)
		return fmt.Errorf("failed to register worker %s: %w", w.id, err)
	}
	w.logger.Info(w.ctx, "Worker registered", "worker_id", w.id, "name", w.name)
	return nil
}

// Heartbeat sends periodic heartbeat messages to the dispatcher
func (w *Worker) Heartbeat() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			w.logger.Debug(w.ctx, "Heartbeat stopped due to context cancellation", "worker_id", w.id)
			return
		case <-ticker.C:
			if !w.nc.IsConnected() {
				w.logger.Error(w.ctx, "Cannot send heartbeat: NATS not connected", "worker_id", w.id)
				continue
			}
			msg := struct {
				WorkerID string `json:"worker_id"`
				Status   string `json:"status"`
			}{
				WorkerID: w.id,
				Status:   "heartbeat",
			}
			data, err := json.Marshal(msg)
			if err != nil {
				w.logger.Error(w.ctx, "Failed to marshal heartbeat message", "worker_id", w.id, "error", err)
				continue
			}
			if err := w.nc.Publish("dispatcher.worker.heartbeat", data); err != nil {
				w.logger.Error(w.ctx, "Failed to send heartbeat", "worker_id", w.id, "error", err)
			} else {
				w.logger.Debug(w.ctx, "Sent heartbeat", "worker_id", w.id)
			}
		}
	}
}

// Deregister sends a deregistration message to the dispatcher
func (w *Worker) Deregister() error {
	if !w.nc.IsConnected() {
		w.logger.Error(w.ctx, "Cannot deregister: NATS not connected", "worker_id", w.id)
		return fmt.Errorf("cannot deregister worker %s: NATS not connected", w.id)
	}
	msg := struct {
		WorkerID string `json:"worker_id"`
		Status   string `json:"status"`
	}{
		WorkerID: w.id,
		Status:   "deregistered",
	}
	data, err := json.Marshal(msg)
	if err != nil {
		w.logger.Error(w.ctx, "Failed to marshal deregistration message", "worker_id", w.id, "error", err)
		return fmt.Errorf("failed to marshal deregistration message: %w", err)
	}
	if err := w.nc.Publish("dispatcher.worker.deregister", data); err != nil {
		w.logger.Error(w.ctx, "Failed to publish deregistration message", "worker_id", w.id, "error", err)
		return fmt.Errorf("failed to deregister worker %s: %w", w.id, err)
	}
	w.logger.Info(w.ctx, "Worker deregistered", "worker_id", w.id)
	return nil
}

// Start begins worker operations
func (w *Worker) Start() error {
	w.logger.Debug(w.ctx, "Starting worker", "worker_id", w.id)
	if !w.nc.IsConnected() {
		w.logger.Error(w.ctx, "Cannot start worker: NATS not connected", "worker_id", w.id)
		return fmt.Errorf("cannot start worker %s: NATS not connected", w.id)
	}
	if err := w.Register(); err != nil {
		w.logger.Error(w.ctx, "Worker failed to start", "worker_id", w.id, "error", err)
		return err
	}
	go w.Heartbeat()
	w.logger.Debug(w.ctx, "Worker started, heartbeat goroutine launched", "worker_id", w.id)
	return nil
}

// Stop shuts down the worker gracefully
func (w *Worker) Stop() error {
	w.logger.Debug(w.ctx, "Stopping worker", "worker_id", w.id)
	w.cancel()
	if err := w.Deregister(); err != nil {
		w.logger.Error(w.ctx, "Failed to deregister during stop", "worker_id", w.id, "error", err)
		return err
	}
	w.logger.Debug(w.ctx, "Worker stopped", "worker_id", w.id)
	return nil
}