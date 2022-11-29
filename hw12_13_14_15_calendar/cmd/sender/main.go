package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	sqlstorage "github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/storage/sql"

	internalconfig "github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/config"
	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/rabbitmq"
	internalsender "github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/sender"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "./../../configs/sender_config.env", "Path to configuration file")
}

const (
	SenderLogFile  = "/opt/sender/logs/sender.log"
	SenderLogLevel = "info"
)

func main() {
	flag.Parse()
	config := internalconfig.NewConfig(configPath)
	logFile, _ := os.OpenFile(SenderLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	defer logFile.Close()
	logg := logger.New(SenderLogLevel, logFile)

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

	sender := internalsender.New(storage, logg, rabbit)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	doneCh := make(chan struct{})
	go func() {
		<-ctx.Done()

		_, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := sender.Shutdown(); err != nil {
			logg.Error(fmt.Sprintf("failed to stop sender %s", err))
		}
		doneCh <- struct{}{}
	}()

	go func() {
		logg.Info("sender is running...")
		if err := sender.Run(ctx); err != nil {
			logg.Error(fmt.Sprintf("failed to start sender: %s", err))
		}
	}()
	<-doneCh
}
