// Package messaging provides an optional Kafka producer. Vidhya Service does
// not publish or consume any Kafka events today - no other service depends
// on it yet. This is wired up only so that when cross-service events are
// needed (e.g. notifying a future "notifications" service), the team can
// flip KAFKA_ENABLED=true and start producing without new plumbing.
package messaging

import (
	"context"
	"fmt"

	kafka "github.com/segmentio/kafka-go"

	"vidhya-service/src/config"
	"vidhya-service/src/logger"
)

// Producer wraps a Kafka writer for a single topic.
type Producer struct {
	writer *kafka.Writer
}

// Connect returns a Producer if Kafka is enabled and reachable, or (nil, nil)
// when disabled. Callers must treat messaging as best-effort in phase 1.
func Connect(ctx context.Context, cfg config.KafkaConfig, topic string) (*Producer, error) {
	if !cfg.Enabled {
		logger.Log.Info("", "Kafka is disabled (KAFKA_ENABLED=false) - skipping connection")
		return nil, nil
	}

	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	logger.Log.Info("", "Kafka producer configured for brokers=%v topic=%s", cfg.Brokers, topic)
	return &Producer{writer: writer}, nil
}

// Publish sends a single message. Safe to call on a nil *Producer (no-op),
// so services can call it unconditionally without checking whether Kafka is
// enabled.
func (p *Producer) Publish(ctx context.Context, key, value []byte) error {
	if p == nil || p.writer == nil {
		return nil
	}
	if err := p.writer.WriteMessages(ctx, kafka.Message{Key: key, Value: value}); err != nil {
		return fmt.Errorf("messaging.Publish: %w", err)
	}
	return nil
}

// Close closes the underlying writer. Safe to call on a nil *Producer.
func (p *Producer) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}
