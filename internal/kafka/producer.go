package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafkago.Writer
}

func NewProducer(broker, topic string) *Producer {
	writer := &kafkago.Writer{
		Addr: kafkago.TCP(broker),

		Topic: topic,

		Balancer: &kafkago.LeastBytes{},

		WriteTimeout: 10 * time.Second,

		ReadTimeout: 10 * time.Second,

		RequiredAcks: kafkago.RequireAll,

		AllowAutoTopicCreation: true,
	}

	slog.Info("kafka producer: connected", "broker", broker, "topic", topic)
	return &Producer{writer: writer}
}

func (p *Producer) PublishCheckJob(ctx context.Context, job CheckJob) error {
	value, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal check job: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafkago.Message{
		Key: []byte(job.MonitorID),

		Value: value,
	})
	if err != nil {
		return fmt.Errorf("write check job to kafka: %w", err)
	}

	slog.Debug("kafka producer: published check job", "monitor", job.MonitorName)
	return nil
}

func (p *Producer) PublishCheckResult(ctx context.Context, result CheckResult) error {
	value, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal check result: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(result.MonitorID),
		Value: value,
	})
	if err != nil {
		return fmt.Errorf("write check result to kafka: %w", err)
	}

	slog.Debug("kafka producer: published check result", "monitor_id", result.MonitorID)
	return nil
}

func (p *Producer) Close() error {
	slog.Info("kafka producer: closing writer")
	return p.writer.Close()
}
