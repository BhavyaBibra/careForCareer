package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"careergps/config"
	"careergps/internal/interfaces/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config load failed: %v\n", err)
		os.Exit(1)
	}

	log := buildWorkerLogger(cfg.LogLevel)
	defer log.Sync()

	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	}

	srv := asynq.NewServer(redisOpt, asynq.Config{
		Queues: map[string]int{
			worker.QueueCritical: 10,
			worker.QueueHigh:     7,
			worker.QueueDefault:  5,
			worker.QueueLow:      2,
		},
		Concurrency: 10,
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			log.Error("job failed",
				zap.String("type", task.Type()),
				zap.Error(err),
			)
		}),
	})

	mux := asynq.NewServeMux()
	mux.HandleFunc(worker.TaskParseResumePDF,    worker.HandleParseResumePDF(log))
	mux.HandleFunc(worker.TaskExtractJD,         worker.HandleExtractJD(log))
	mux.HandleFunc(worker.TaskRunGapAnalysis,    worker.HandleRunGapAnalysis(log))

	go func() {
		log.Info("worker starting")
		if err := srv.Run(mux); err != nil {
			log.Fatal("worker error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("worker shutting down")
	srv.Shutdown()
	log.Info("worker stopped")
}

func buildWorkerLogger(level string) *zap.Logger {
	cfg := zap.NewProductionConfig()
	log, _ := cfg.Build()
	return log
}
