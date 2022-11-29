package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	internalscheduler "github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/scheduler"
	sqlstorage "github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/storage/sql"

	internalconfig "github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/config"
	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/rabbitmq"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "./../../configs/scheduler_config.env", "Path to configuration file")
}

const (
	SchedulerLogFile  = "/opt/scheduler/logs/scheduler.log"
	SchedulerLogLevel = "info"
)

func main() {
	flag.Parse()
	config := internalconfig.NewConfig(configPath)

	logFile, _ := os.OpenFile(SchedulerLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	defer logFile.Close()
	logg := logger.New(SchedulerLogLevel, logFile)

	storage := sqlstorage.New(config)
	err := storage.Connect()
	if err != nil {
		logg.Fatal(fmt.Sprintf("can not connect to storage: %s", err))
	}

	rabbit := rabbitmq.New(config)
	err = rabbit.Connect()
	if err != nil {
		logg.Fatal(fmt.Sprintf("can not connect to RabbitMQ: %s", err))
	}
	defer func() {
		err := rabbit.Close()
		if err != nil {
			logg.Error("failed to close connection RabbitMQ")
		}
	}()

	scheduler := internalscheduler.New(config, storage, logg, rabbit)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	doneCh := make(chan struct{})
	go func() {
		<-ctx.Done()

		if err := scheduler.Shutdown(); err != nil {
			logg.Error(fmt.Sprintf("failed to stop scheduler %s", err))
		}
		doneCh <- struct{}{}
	}()

	go func() {
		logg.Info("scheduler is running...")
		if err := scheduler.Run(ctx); err != nil {
			logg.Info(fmt.Sprintf("failed to start scheduler: %s", err))
		}
	}()
	<-doneCh
}
