package rabbitmq

import (
	"context"
	"encoding/json"

	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Rabbit struct {
	dsn       string
	queueName string
	conn      *amqp.Connection
	channel   *amqp.Channel
}

func New(config config.Config) *Rabbit {
	return &Rabbit{
		dsn:       config.AMQPAddress,
		queueName: config.AMQPQueueName,
	}
}

func (r *Rabbit) Connect() error {
	conn, err := amqp.Dial(r.dsn)
	if err != nil {
		return err
	}
	r.conn = conn

	amqpCh, err := conn.Channel()
	if err != nil {
		return err
	}
	r.channel = amqpCh

	_, err = amqpCh.QueueDeclare(
		r.queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *Rabbit) Close() error {
	err := r.conn.Close()
	if err != nil {
		return err
	}
	err = r.channel.Close()
	if err != nil {
		return err
	}
	return nil
}

func (r *Rabbit) Publish(ctx context.Context, data interface{}) error {
	encodedData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return r.channel.PublishWithContext(
		ctx,
		"",
		r.queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        encodedData,
		},
	)
}

func (r *Rabbit) Consume() (<-chan amqp.Delivery, error) {
	msgs, err := r.channel.Consume(
		r.queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return msgs, nil
}
