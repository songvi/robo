package dispatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/fx"

	"github.com/songvi/robo/config"
	"github.com/songvi/robo/logger"
	"github.com/songvi/robo/models"
)

// WorkerRegistrationMessage defines the structure of worker registration messages
type WorkerRegistrationMessage struct {
	WorkerID     string   `json:"worker_id"`
	Name         string   `json:"name"`
	Capabilities []string `json:"capabilities"`
	Status       string   `json:"status"`
}

// Dispatcher defines the interface for the dispatcher service
type Dispatcher interface {
	Publish(ctx context.Context, subject string, data []byte) error
	Subscribe(ctx context.Context, subject string) (<-chan *nats.Msg, error)
	GetActiveWorkers() []models.Worker
	DispatchJob(ctx context.Context, job *models.Job) error
}

// dispatcherImpl is the implementation of the Dispatcher interface
type dispatcherImpl struct {
	nc            *nats.Conn
	logger        logger.Logger
	workers       map[string]models.Worker
	workerMu      sync.RWMutex
	lastHeartbeat map[string]time.Time
	heartbeatMu   sync.RWMutex
}

// NewDispatcher creates a new Dispatcher instance
func NewDispatcher(lc fx.Lifecycle, configService config.ConfigService, logger logger.Logger) (Dispatcher, error) {
	config := configService.GetConfig()
	broker := config.Broker
	if broker == "" {
		broker = "nats://localhost:4222"
	}

	// Connect to NATS
	nc, err := nats.Connect(broker)
	if err != nil {
		logger.Error(context.Background(), "Failed to connect to NATS", "broker", broker, "error", err)
		return nil, err
	}

	d := &dispatcherImpl{
		nc:            nc,
		logger:        logger,
		workers:       make(map[string]models.Worker),
		lastHeartbeat: make(map[string]time.Time),
	}

	// Start worker registration and heartbeat handling
	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			d.logger.Info(ctx, "Dispatcher connected to NATS", "broker", broker)
			if err := d.startWorkerManagement(ctx); err != nil {
				return err
			}
			return nil
		},
		OnStop: func(context.Context) error {
			d.logger.Info(ctx, "Closing NATS connection")
			cancel()
			nc.Close()
			return nil
		},
	})

	return d, nil
}

// DispatchJob sends a job to an active worker
func (d *dispatcherImpl) DispatchJob(ctx context.Context, job *models.Job) error {
	// Get active workers
	workers := d.GetActiveWorkers()
	if len(workers) == 0 {
		d.logger.Error(ctx, "No active workers available to dispatch job", "job_uuid", job.UUID)
		return fmt.Errorf("no active workers available")
	}

	// Select a worker randomly (modify for a different strategy if needed)
	worker := workers[rand.Intn(len(workers))]
	job.WorkerID = worker.UUID

	// Serialize job to JSON
	data, err := json.Marshal(job)
	if err != nil {
		d.logger.Error(ctx, "Failed to marshal job", "job_uuid", job.UUID, "error", err)
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Publish job to worker-specific subject
	subject := fmt.Sprintf("dispatcher.job.%s", worker.UUID)
	if err := d.Publish(ctx, subject, data); err != nil {
		d.logger.Error(ctx, "Failed to dispatch job", "job_uuid", job.UUID, "worker_id", worker.UUID, "error", err)
		return fmt.Errorf("failed to dispatch job: %w", err)
	}

	d.logger.Info(ctx, "Dispatched job to worker", "job_uuid", job.UUID, "worker_id", worker.UUID, "job_name", job.Name)
	return nil
}

// startWorkerManagement sets up subscriptions for worker registration, heartbeats, and deregistration
func (d *dispatcherImpl) startWorkerManagement(ctx context.Context) error {
	// Subscribe to worker registration
	regCh, err := d.Subscribe(ctx, "dispatcher.worker.register")
	if err != nil {
		return err
	}
	go d.handleRegistrations(ctx, regCh)

	// Subscribe to worker heartbeats
	hbCh, err := d.Subscribe(ctx, "dispatcher.worker.heartbeat")
	if err != nil {
		return err
	}
	go d.handleHeartbeats(ctx, hbCh)

	// Subscribe to worker deregistration
	derCh, err := d.Subscribe(ctx, "dispatcher.worker.deregister")
	if err != nil {
		return err
	}
	go d.handleDeregistrations(ctx, derCh)

	// Start heartbeat cleanup
	go d.cleanupInactiveWorkers(ctx)

	return nil
}

// handleRegistrations processes worker registration messages
func (d *dispatcherImpl) handleRegistrations(ctx context.Context, regCh <-chan *nats.Msg) {
	for msg := range regCh {
		var regMsg WorkerRegistrationMessage
		if err := json.Unmarshal(msg.Data, &regMsg); err != nil {
			d.logger.Error(ctx, "Failed to unmarshal registration message", "error", err)
			continue
		}
		if regMsg.Status != "registered" {
			continue
		}

		worker := models.Worker{
			Name: regMsg.Name,
			UUID: regMsg.WorkerID,
		}
		d.workerMu.Lock()
		d.workers[regMsg.WorkerID] = worker
		d.workerMu.Unlock()

		d.heartbeatMu.Lock()
		d.lastHeartbeat[regMsg.WorkerID] = time.Now()
		d.heartbeatMu.Unlock()

		d.logger.Info(ctx, "Worker registered", "worker_id", regMsg.WorkerID, "name", regMsg.Name, "capabilities", regMsg.Capabilities)
	}
}

// handleHeartbeats processes worker heartbeat messages
func (d *dispatcherImpl) handleHeartbeats(ctx context.Context, hbCh <-chan *nats.Msg) {
	for msg := range hbCh {
		var hbMsg WorkerRegistrationMessage
		if err := json.Unmarshal(msg.Data, &hbMsg); err != nil {
			d.logger.Error(ctx, "Failed to unmarshal heartbeat message", "error", err)
			continue
		}
		if hbMsg.Status != "heartbeat" {
			continue
		}

		d.heartbeatMu.Lock()
		d.lastHeartbeat[hbMsg.WorkerID] = time.Now()
		d.heartbeatMu.Unlock()

		d.logger.Info(ctx, "Received heartbeat", "worker_id", hbMsg.WorkerID)
	}
}

// handleDeregistrations processes worker deregistration messages
func (d *dispatcherImpl) handleDeregistrations(ctx context.Context, derCh <-chan *nats.Msg) {
	for msg := range derCh {
		var derMsg WorkerRegistrationMessage
		if err := json.Unmarshal(msg.Data, &derMsg); err != nil {
			d.logger.Error(ctx, "Failed to unmarshal deregistration message", "error", err)
			continue
		}
		if derMsg.Status != "deregistered" {
			continue
		}

		d.workerMu.Lock()
		delete(d.workers, derMsg.WorkerID)
		d.workerMu.Unlock()

		d.heartbeatMu.Lock()
		delete(d.lastHeartbeat, derMsg.WorkerID)
		d.heartbeatMu.Unlock()

		d.logger.Info(ctx, "Worker deregistered", "worker_id", derMsg.WorkerID)
	}
}

// cleanupInactiveWorkers removes workers that haven't sent heartbeats
func (d *dispatcherImpl) cleanupInactiveWorkers(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.heartbeatMu.Lock()
			now := time.Now()
			for workerID, lastHB := range d.lastHeartbeat {
				if now.Sub(lastHB) > 15*time.Second {
					d.workerMu.Lock()
					delete(d.workers, workerID)
					d.workerMu.Unlock()
					delete(d.lastHeartbeat, workerID)
					d.logger.Info(ctx, "Removed inactive worker", "worker_id", workerID)
				}
			}
			d.heartbeatMu.Unlock()
		}
	}
}

// Publish publishes a message to the specified subject
func (d *dispatcherImpl) Publish(ctx context.Context, subject string, data []byte) error {
	if err := d.nc.Publish(subject, data); err != nil {
		d.logger.Error(ctx, "Failed to publish message", "subject", subject, "error", err)
		return err
	}
	d.logger.Info(ctx, "Published message", "subject", subject)
	return nil
}

// Subscribe subscribes to a subject and returns a channel for messages
func (d *dispatcherImpl) Subscribe(ctx context.Context, subject string) (<-chan *nats.Msg, error) {
	msgCh := make(chan *nats.Msg, 64)
	sub, err := d.nc.ChanSubscribe(subject, msgCh)
	if err != nil {
		d.logger.Error(ctx, "Failed to subscribe to subject", "subject", subject, "error", err)
		return nil, err
	}

	// Handle unsubscription on context cancellation
	go func() {
		<-ctx.Done()
		d.logger.Info(ctx, "Unsubscribing from subject", "subject", subject)
		sub.Unsubscribe()
		close(msgCh)
	}()

	d.logger.Info(ctx, "Subscribed to subject", "subject", subject)
	return msgCh, nil
}

// GetActiveWorkers returns the list of active workers
func (d *dispatcherImpl) GetActiveWorkers() []models.Worker {
	d.workerMu.RLock()
	defer d.workerMu.RUnlock()
	workers := make([]models.Worker, 0, len(d.workers))
	for _, w := range d.workers {
		workers = append(workers, w)
	}
	return workers
}

// Module defines the Fx module for the Dispatcher service
var Module = fx.Module(
	"dispatcher",
	fx.Provide(NewDispatcher),
)
