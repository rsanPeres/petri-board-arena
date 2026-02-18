package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/petri-board-arena/internal/infrastructure/messaging"
	"github.com/petri-board-arena/internal/infrastructure/projector"
)

type Consumer struct {
	cfg       ConsumerConfig
	reader    *kafka.Reader
	dlqWriter *kafka.Writer
	projector *projector.Projector
}

func NewConsumer(cfg ConsumerConfig, projector *projector.Projector) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:         cfg.KafkaBrokers,
		Topic:           cfg.KafkaTopic,
		GroupID:         cfg.KafkaGroupID,
		MinBytes:        cfg.PollMinBytes,
		MaxBytes:        cfg.PollMaxBytes,
		ReadLagInterval: -1,
		MaxWait:         250 * time.Millisecond,
		CommitInterval:  cfg.CommitInterval,
	})

	w := &kafka.Writer{
		Addr:         kafka.TCP(cfg.KafkaBrokers...),
		Topic:        cfg.KafkaDLQ,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireOne,
	}

	return &Consumer{
		cfg:       cfg,
		reader:    r,
		dlqWriter: w,
		projector: projector,
	}
}

func (c *Consumer) Close() error {
	_ = c.reader.Close()
	_ = c.dlqWriter.Close()
	return nil
}

func (c *Consumer) Run(ctx context.Context) error {
	log.Printf("[worker] consuming topic=%s group=%s dlq=%s brokers=%v",
		c.cfg.KafkaTopic, c.cfg.KafkaGroupID, c.cfg.KafkaDLQ, c.cfg.KafkaBrokers)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		readCtx, cancel := context.WithTimeout(ctx, c.cfg.ReadTimeout)
		msg, err := c.reader.FetchMessage(readCtx)
		cancel()

		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				continue
			}
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}

		if err := c.processWithRetry(ctx, msg); err != nil {
			_ = c.sendDLQ(ctx, msg, err)
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}

func (c *Consumer) processWithRetry(ctx context.Context, msg kafka.Message) error {
	var lastErr error

	for attempt := 1; attempt <= c.cfg.MaxRetries; attempt++ {
		var ev messaging.EventEnvelope
		if err := json.Unmarshal(msg.Value, &ev); err != nil {
			return err
		}
		if ev.OccurredAt.IsZero() {
			ev.OccurredAt = time.Now().UTC()
		}

		if err := c.projector.Apply(ctx, ev); err == nil {
			return nil
		} else {
			lastErr = err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(c.cfg.RetryBackoff * time.Duration(attempt)):
		}
	}
	return lastErr
}

func (c *Consumer) sendDLQ(ctx context.Context, msg kafka.Message, cause error) error {
	key := msg.Key

	dlqPayload := map[string]any{
		"originalTopic":     c.cfg.KafkaTopic,
		"originalPartition": msg.Partition,
		"originalOffset":    msg.Offset,
		"error":             cause.Error(),
		"ts":                time.Now().UTC().Format(time.RFC3339Nano),
		"value":             json.RawMessage(msg.Value),
	}

	b, _ := json.Marshal(dlqPayload)

	return c.dlqWriter.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: b,
		Time:  time.Now(),
	})
}
