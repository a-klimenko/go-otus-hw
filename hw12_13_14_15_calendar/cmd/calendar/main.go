package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/app"
	internalconfig "github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/config"
	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/logger"
	internalgrpc "github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/server/grpc"
	internalhttp "github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/server/http"
	memorystorage "github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./../../configs/config.env", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config := internalconfig.NewConfig(configFile)

	logFile, err := os.OpenFile(config.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer func() {
		err := logFile.Close()
		if err != nil {
			log.Fatalf("can not close log file: %v", err)
		}
	}()
	logg := logger.New(config.LogLevel, logFile)

	var storage app.Storage
	switch config.StorageType {
	case internalconfig.SQLStorageType:
		storage = sqlstorage.New(config)
	case internalconfig.InmemoryStorageType:
		storage = memorystorage.New()
	default:
		log.Fatal("unregistered storage type") //nolint:gocritic
	}

	err = storage.Connect(context.Background())
	if err != nil {
		logg.Error(fmt.Sprintf("can not connect to storage: %s", err))
	}

	calendar := app.New(logg, storage)

	httpServer := internalhttp.NewServer(config.Host, config.HTTPPort, logg, calendar)
	grpcServer := internalgrpc.NewServer(config.Host, config.GRPCPort, logg, calendar)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	doneCh := make(chan struct{})
	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := httpServer.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
		if err := grpcServer.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
		doneCh <- struct{}{}
	}()

	logg.Info("calendar is running...")

	if err := httpServer.Start(ctx); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
		os.Exit(1)
	}
	if err := grpcServer.Start(ctx); err != nil {
		logg.Error("failed to start grpc server: " + err.Error())
		cancel()
		os.Exit(1)
	}
	<-doneCh
}
