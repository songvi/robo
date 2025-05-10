package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/songvi/robo/config"
	"github.com/songvi/robo/dispatcher"
	"github.com/songvi/robo/generator"
	"github.com/songvi/robo/job"
	"github.com/songvi/robo/logger"
)

// CustomFxLogger adapts logger.Logger to fxevent.Logger
type CustomFxLogger struct {
	logger logger.Logger
}

func (l *CustomFxLogger) LogEvent(event fxevent.Event) {
	ctx := context.Background()
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		l.logger.Debug(ctx, "Fx OnStart hook executing", "callee", e.FunctionName, "caller", e.CallerName)
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			l.logger.Error(ctx, "Fx OnStart hook failed", "callee", e.FunctionName, "caller", e.CallerName, "error", e.Err)
		} else {
			l.logger.Debug(ctx, "Fx OnStart hook executed", "callee", e.FunctionName, "caller", e.CallerName)
		}
	case *fxevent.OnStopExecuting:
		l.logger.Debug(ctx, "Fx OnStop hook executing", "callee", e.FunctionName, "caller", e.CallerName)
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			l.logger.Error(ctx, "Fx OnStop hook failed", "callee", e.FunctionName, "caller", e.CallerName, "error", e.Err)
		} else {
			l.logger.Debug(ctx, "Fx OnStart hook executed", "callee", e.FunctionName, "caller", e.CallerName)
		}
	case *fxevent.Provided:
		l.logger.Debug(ctx, "Fx provided", "constructor", e.ConstructorName)
	case *fxevent.Invoked:
		if e.Err != nil {
			l.logger.Error(ctx, "Fx invoke failed", "function", e.FunctionName, "error", e.Err)
		} else {
			l.logger.Debug(ctx, "Fx invoked", "function", e.FunctionName)
		}
	case *fxevent.Started:
		l.logger.Info(ctx, "Fx application started")
	case *fxevent.Stopped:
		l.logger.Info(ctx, "Fx application stopped")
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			l.logger.Error(ctx, "Fx logger initialization failed", "error", e.Err)
		} else {
			l.logger.Debug(ctx, "Fx logger initialized")
		}
	}
}

func main() {
	app := fx.New(
		fx.WithLogger(func(logger logger.Logger) fxevent.Logger {
			return &CustomFxLogger{logger: logger}
		}),
		logger.ProvideLogger(),
		config.ProvideConfigService(),
		generator.Module,
		dispatcher.Module,
		fx.Invoke(func(d dispatcher.Dispatcher, logger logger.Logger) {
			ctx := context.Background()
			logger.Info(ctx, "Invoking Dispatcher lifecycle")

			go func() {
				time.Sleep(5 * time.Second) // Wait for worker to register
				job := &job.Job{
					UUID:      uuid.New().String(),
					Name:      "test-job",
					InputData: json.RawMessage(`{"task":"process_file"}`),
					Status:    "pending",
				}
				// Retry up to 3 times
				for attempt := 1; attempt <= 5; attempt++ {
					workers := d.GetActiveWorkers()
					logger.Debug(ctx, "Attempting to dispatch job", "job_uuid", job.UUID, "attempt", attempt, "active_workers", len(workers))
					if err := d.DispatchJob(ctx, job); err != nil {
						logger.Error(ctx, "Failed to dispatch test job", "job_uuid", job.UUID, "attempt", attempt, "error", err)
						if attempt < 5 {
							time.Sleep(2 * time.Second)
							continue
						}
					} else {
						logger.Info(ctx, "Test job dispatched successfully", "job_uuid", job.UUID)
						break
					}
				}
			}()

		}),
		fx.Invoke(func(lc fx.Lifecycle, logger logger.Logger) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					logger.Debug(ctx, "Fx application starting")
					return nil
				},
				OnStop: func(ctx context.Context) error {
					logger.Debug(ctx, "Fx application stopping")
					return nil
				},
			})
		}),
	)

	if err := app.Err(); err != nil {
		logger := logger.NewSlogLogger()
		logger.Error(context.Background(), "Failed to initialize Fx app", "error", err)
		return
	}

	app.Run()
}
