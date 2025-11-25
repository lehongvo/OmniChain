package messagequeue

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/streadway/amqp"

	"github.com/onichange/pos-system/pkg/logger"
)

// RabbitMQ represents a RabbitMQ connection
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	logger  *logger.Logger
}

// NewRabbitMQ creates a new RabbitMQ connection
func NewRabbitMQ(url string, log *logger.Logger) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &RabbitMQ{
		conn:    conn,
		channel: channel,
		logger:  log,
	}, nil
}

// DeclareExchange declares an exchange
func (r *RabbitMQ) DeclareExchange(name, kind string) error {
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
func (r *RabbitMQ) DeclareQueue(name string, durable bool) (amqp.Queue, error) {
	return r.channel.QueueDeclare(
		name,    // name
		durable, // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
}

// DeclareDeadLetterQueue declares a dead letter queue for failed messages
func (r *RabbitMQ) DeclareDeadLetterQueue(name string) (amqp.Queue, error) {
	return r.channel.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		amqp.Table{
			"x-message-ttl": int32(7 * 24 * time.Hour / time.Millisecond), // 7 days
		},
	)
}

// Publish publishes a message to an exchange
func (r *RabbitMQ) Publish(exchange, routingKey string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return r.channel.Publish(
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // Make message persistent
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
}

// Consume consumes messages from a queue
func (r *RabbitMQ) Consume(queue, consumer string, handler func(amqp.Delivery) error) error {
	msgs, err := r.channel.Consume(
		queue,    // queue
		consumer, // consumer
		false,    // auto-ack (manual ack)
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for msg := range msgs {
			if err := handler(msg); err != nil {
				r.logger.Errorf("Error processing message: %v", err)
				// Nack and requeue
				msg.Nack(false, true)
			} else {
				// Ack message
				msg.Ack(false)
			}
		}
	}()

	return nil
}

// Close closes the RabbitMQ connection
func (r *RabbitMQ) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

// Event represents a domain event
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// PublishEvent publishes a domain event
func (r *RabbitMQ) PublishEvent(eventType, routingKey string, data map[string]interface{}) error {
	event := Event{
		ID:        fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		Type:      eventType,
		Source:    "onichange-pos",
		Timestamp: time.Now(),
		Data:      data,
	}

	return r.Publish("events", routingKey, event)
}
