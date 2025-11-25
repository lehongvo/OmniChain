package messaging

import (
	"fmt"
	"time"

	"github.com/streadway/amqp"

	"github.com/onichange/pos-system/pkg/logger"
)

// RabbitMQClient wraps RabbitMQ connection and channel
type RabbitMQClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	logger  *logger.Logger
}

// NewRabbitMQClient creates a new RabbitMQ client
func NewRabbitMQClient(url string, log *logger.Logger) (*RabbitMQClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	log.Info("Connected to RabbitMQ")

	return &RabbitMQClient{
		conn:    conn,
		channel: channel,
		logger:  log,
	}, nil
}

// DeclareExchange declares an exchange
func (r *RabbitMQClient) DeclareExchange(name, kind string) error {
	return r.channel.ExchangeDeclare(
		name,  // name
		kind,  // kind (direct, topic, fanout, headers)
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
}

// DeclareQueue declares a queue
func (r *RabbitMQClient) DeclareQueue(name string) (amqp.Queue, error) {
	return r.channel.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
}

// BindQueue binds a queue to an exchange
func (r *RabbitMQClient) BindQueue(queue, key, exchange string) error {
	return r.channel.QueueBind(
		queue,    // queue name
		key,      // routing key
		exchange, // exchange
		false,
		nil,
	)
}

// Publish publishes a message to an exchange
func (r *RabbitMQClient) Publish(exchange, key string, body []byte) error {
	return r.channel.Publish(
		exchange, // exchange
		key,      // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // Make message persistent
			Timestamp:    time.Now(),
		},
	)
}

// Consume consumes messages from a queue
func (r *RabbitMQClient) Consume(queue, consumer string) (<-chan amqp.Delivery, error) {
	return r.channel.Consume(
		queue,    // queue
		consumer, // consumer
		false,    // auto-ack
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)
}

// Close closes the connection
func (r *RabbitMQClient) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
