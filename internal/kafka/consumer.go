package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	kafkago "github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafkago.Reader
}

func NewJobConsumer(broker, groupID string) *Consumer {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: []string{broker},

		Topic: CheckJobsTopic,

		GroupID: groupID,

		MinBytes: 10e3, // 10 KB
		MaxBytes: 10e6, // 10 MB

		CommitInterval: 0,
	})

	slog.Info("kafka consumer: reader created", "topic", CheckJobsTopic, "group", groupID)
	return &Consumer{reader: reader}
}

func NewResultConsumer(broker string) *Consumer {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:        []string{broker},
		Topic:          CheckResultsTopic,
		GroupID:        "pulsedb-result-writers",
		MinBytes:       10e3,
		MaxBytes:       10e6,
		CommitInterval: 0,
	})

	slog.Info("kafka consumer: result reader created", "topic", CheckResultsTopic)
	return &Consumer{reader: reader}
}

func (c *Consumer) ReadCheckJob(ctx context.Context) (CheckJob, kafkago.Message, error) {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return CheckJob{}, kafkago.Message{}, fmt.Errorf("fetch message: %w", err)
	}

	var job CheckJob
	if err := json.Unmarshal(msg.Value, &job); err != nil {
		_ = c.reader.CommitMessages(ctx, msg)
		return CheckJob{}, kafkago.Message{}, fmt.Errorf("unmarshal check job: %w", err)
	}

	return job, msg, nil
}

func (c *Consumer) ReadCheckResult(ctx context.Context) (CheckResult, kafkago.Message, error) {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return CheckResult{}, kafkago.Message{}, fmt.Errorf("fetch result: %w", err)
	}

	var result CheckResult
	if err := json.Unmarshal(msg.Value, &result); err != nil {
		_ = c.reader.CommitMessages(ctx, msg)
		return CheckResult{}, kafkago.Message{}, fmt.Errorf("unmarshal check result: %w", err)
	}

	return result, msg, nil
}

func (c *Consumer) Commit(ctx context.Context, msg kafkago.Message) error {
	return c.reader.CommitMessages(ctx, msg)
}

func (c *Consumer) Close() error {
	slog.Info("kafka consumer: closing reader")
	return c.reader.Close()
}
