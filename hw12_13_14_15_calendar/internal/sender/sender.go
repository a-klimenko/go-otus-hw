package sender

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/app"
	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/storage"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Sender struct {
	queue   Queue
	storage app.Storage
	logger  app.Logger
}

type Queue interface {
	Consume() (<-chan amqp.Delivery, error)
	Close() error
}

func New(storage app.Storage, logger app.Logger, queue Queue) *Sender {
	return &Sender{
		storage: storage,
		logger:  logger,
		queue:   queue,
	}
}

func (s *Sender) Run(ctx context.Context) error {
	msgs, err := s.queue.Consume()
	if err != nil {
		return fmt.Errorf("failed to consume amqp, %w", err)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}
			s.logger.Info("receive notification from a queue")

			notification := &storage.Notification{}

			if err := json.Unmarshal(msg.Body, notification); err != nil {
				s.logger.Error(fmt.Sprintf("event notification unmarshal failed %s", err))
				return nil
			}
			s.logger.Info(notification.String())

			if err := msg.Ack(false); err != nil {
				s.logger.Error(fmt.Sprintf("ack failed %s", err))
			}

			if err := s.storage.ChangeNotifyStatus(ctx, notification.EventID); err != nil {
				s.logger.Error(fmt.Sprintf("notify failed %s", err))
			}
		}
	}
}

func (s *Sender) Shutdown() error {
	s.logger.Info("sender is shutting down")

	return s.queue.Close()
}
