package messaging

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"

	"github.com/onichange/pos-system/pkg/logger"
)

// KafkaProducer wraps Kafka producer
type KafkaProducer struct {
	producer sarama.SyncProducer
	logger   *logger.Logger
}

// KafkaConsumer wraps Kafka consumer
type KafkaConsumer struct {
	consumer sarama.ConsumerGroup
	logger   *logger.Logger
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(brokers []string, log *logger.Logger) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Compression = sarama.CompressionSnappy

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	log.Info("Kafka producer created")

	return &KafkaProducer{
		producer: producer,
		logger:   log,
	}, nil
}

// Publish publishes a message to a topic
func (k *KafkaProducer) Publish(topic string, key []byte, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	partition, offset, err := k.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	k.logger.Debugf("Message sent to topic %s, partition %d, offset %d", topic, partition, offset)
	return nil
}

// Close closes the producer
func (k *KafkaProducer) Close() error {
	return k.producer.Close()
}

// NewKafkaConsumer creates a new Kafka consumer
func NewKafkaConsumer(brokers []string, groupID string, log *logger.Logger) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	log.Info("Kafka consumer created")

	return &KafkaConsumer{
		consumer: consumer,
		logger:   log,
	}, nil
}

// Consume consumes messages from topics
func (k *KafkaConsumer) Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	return k.consumer.Consume(ctx, topics, handler)
}

// Close closes the consumer
func (k *KafkaConsumer) Close() error {
	return k.consumer.Close()
}
