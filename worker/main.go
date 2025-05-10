package main

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/songvi/robo/config"
	"github.com/songvi/robo/logger"
)

// ProvideNATS provides a NATS connection using Config.Broker
func ProvideNATS(lc fx.Lifecycle, configService config.ConfigService, logger logger.Logger) (*nats.Conn, error) {
	ctx := context.Background()
	logger.Debug(ctx, "Initializing NATS connection")
	config := configService.GetConfig()
	broker := config.Broker
	if broker == "" {
		broker = "nats://localhost:4222"
		logger.Info(ctx, "Using default NATS broker", "broker", broker)
	} else {
		logger.Info(ctx, "Using configured NATS broker", "broker", broker)
	}

	// Connect with timeout and retry
	nc, err := nats.Connect(broker,
		nats.Timeout(5*time.Second),
		nats.MaxReconnects(3),
		nats.ReconnectWait(time.Second),
	)
	if err != nil {
		logger.Error(ctx, "Failed to connect to NATS", "broker", broker, "error", err)
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Verify connection
	if !nc.IsConnected() {
		logger.Error(ctx, "NATS connection is not active", "broker", broker)
		nc.Close()
		return nil, fmt.Errorf("NATS connection is not active")
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info(ctx, "Closing NATS connection")
			nc.Close()
			return nil
		},
	})

	logger.Info(ctx, "Successfully connected to NATS", "broker", broker)
	return nc, nil
}

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
			l.logger.Debug(ctx, "Fx OnStop hook executed", "callee", e.FunctionName, "caller", e.CallerName)
		}
	case *fxevent.Provided:
		l.logger.Debug(ctx, "Fx provided", "constructor", e.ConstructorName)
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
		config.Module,
		fx.Provide(ProvideNATS),
		fx.Provide(NewWorker),
		fx.Invoke(func(w Worker, logger logger.Logger) {
			logger.Debug(context.Background(), "Invoking Worker lifecycle")
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
