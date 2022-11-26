package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/app"
	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/config"
	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/storage"
)

type Queue interface {
	Publish(ctx context.Context, data interface{}) error
	Close() error
}

type Scheduler struct {
	timeout time.Duration
	storage app.Storage
	logger  app.Logger
	queue   Queue
}

func New(config config.Config, storage app.Storage, logger app.Logger, queue Queue) *Scheduler {
	return &Scheduler{
		timeout: config.EventScanTimeout,
		storage: storage,
		logger:  logger,
		queue:   queue,
	}
}

func (s *Scheduler) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.timeout)
	defer ticker.Stop()

	for {
		go func() {
			s.notify(ctx)
			err := s.storage.DeleteOldEvents(ctx)
			if err != nil {
				s.logger.Error(fmt.Sprintf("fail delete old events %s", err))
			}
		}()

		select {
		case <-ctx.Done():
			s.logger.Info("stop scheduler")

			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (s *Scheduler) notify(ctx context.Context) {
	startDate := time.Now().Add(-s.timeout + time.Second)
	endDate := time.Now()

	events, err := s.storage.GetByNotificationPeriod(ctx, startDate, endDate)
	if err != nil {
		s.logger.Error(fmt.Sprintf("fail get event for notify %s", err))
	}

	s.logger.Info(fmt.Sprintf("found %d events to notify", len(events)))

	for _, event := range events {
		notification := storage.Notification{
			EventID:    event.ID,
			EventTitle: event.Title,
			DateTime:   event.NotificationDate,
			UserID:     event.UserID,
		}
		if err := s.queue.Publish(ctx, notification); err != nil {
			s.logger.Error(fmt.Sprintf("fail publish event notification %s", err))
		}
	}
}

func (s *Scheduler) Shutdown() error {
	s.logger.Info("scheduler is shutting down")

	return s.queue.Close()
}
