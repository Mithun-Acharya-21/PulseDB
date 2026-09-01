package checker

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"myproject/internal/kafka"
	"myproject/internal/metrics"
	"myproject/internal/monitor"
)

const (
	numWorkers    = 5
	scheduleEvery = 10 * time.Second
	probeTimeout  = 10 * time.Second
)

type Checker struct {
	repo          monitor.Repository
	jobProducer   *kafka.Producer  // Scheduler → Kafka (CheckJobsTopic)
	resultProducer *kafka.Producer // Workers → Kafka (CheckResultsTopic)
	broker        string
}

func New(repo monitor.Repository, broker string) *Checker {
	return &Checker{
		repo:           repo,
		jobProducer:    kafka.NewProducer(broker, kafka.CheckJobsTopic),
		resultProducer: kafka.NewProducer(broker, kafka.CheckResultsTopic),
		broker:         broker,
	}
}

func (c *Checker) Start(ctx context.Context) {
	var wg sync.WaitGroup

	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			c.runWorker(ctx, workerID)
		}(i)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		c.runResultWriter(ctx)
	}()

	c.scheduler(ctx)

	wg.Wait()

	c.jobProducer.Close()
	c.resultProducer.Close()

	slog.Info("checker: all goroutines stopped cleanly")
}

func (c *Checker) scheduler(ctx context.Context) {
	ticker := time.NewTicker(scheduleEvery)
	defer ticker.Stop()

	c.dispatch(ctx)

	for {
		select {
		case <-ticker.C:
			c.dispatch(ctx)
		case <-ctx.Done():
			slog.Info("checker: scheduler stopping")
			return
		}
	}
}

func (c *Checker) dispatch(ctx context.Context) {
	monitors, err := c.repo.ListDueMonitors(ctx)
	if err != nil {
		slog.Error("checker: list due monitors", "err", err)
		return
	}

	metrics.ActiveMonitors.Set(float64(len(monitors)))
	metrics.WorkerQueueDepth.Set(float64(len(monitors)))

	slog.Info("checker: dispatching jobs", "count", len(monitors))

	for _, m := range monitors {
		job := kafka.CheckJob{
			MonitorID:   m.ID,
			MonitorName: m.Name,
			URL:         m.URL,
		}

		if err := c.jobProducer.PublishCheckJob(ctx, job); err != nil {
			slog.Error("checker: publish job", "monitor", m.Name, "err", err)
		}
	}
}

func (c *Checker) runWorker(ctx context.Context, id int) {
	consumer := kafka.NewJobConsumer(c.broker, kafka.ConsumerGroupID)
	defer consumer.Close()

	slog.Info("checker: worker started", "id", id)

	for {
		job, msg, err := consumer.ReadCheckJob(ctx)
		if err != nil {
			if ctx.Err() != nil {
				slog.Info("checker: worker stopping", "id", id)
				return
			}
			slog.Error("checker: read job", "worker", id, "err", err)
			continue
		}

		slog.Info("checker: worker processing", "id", id, "monitor", job.MonitorName)

		result := c.probe(ctx, job)

		if err := c.resultProducer.PublishCheckResult(ctx, result); err != nil {
			slog.Error("checker: publish result", "worker", id, "err", err)
		}

		if err := consumer.Commit(ctx, msg); err != nil {
			slog.Error("checker: commit job offset", "worker", id, "err", err)
		}
	}
}

func (c *Checker) probe(ctx context.Context, job kafka.CheckJob) kafka.CheckResult {
	probeCtx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	start := time.Now()
	statusCode, errStr := doProbe(probeCtx, job.URL)
	elapsed := time.Since(start)
	ms := int(elapsed.Milliseconds())

	metrics.CheckDuration.WithLabelValues(job.MonitorName).Observe(elapsed.Seconds())
	if errStr != "" {
		metrics.ChecksTotal.WithLabelValues("error").Inc()
	} else {
		metrics.ChecksTotal.WithLabelValues("success").Inc()
	}

	slog.Info("checker: probed",
		"monitor", job.MonitorName,
		"status", statusCode,
		"ms", ms)

	result := kafka.CheckResult{
		MonitorID:  job.MonitorID,
		StatusCode: statusCode,
		ResponseMs: ms,
	}
	if errStr != "" {
		result.Error = &errStr
	}
	return result
}

func (c *Checker) runResultWriter(ctx context.Context) {
	consumer := kafka.NewResultConsumer(c.broker)
	defer consumer.Close()

	slog.Info("checker: result writer started")

	for {
		result, msg, err := consumer.ReadCheckResult(ctx)
		if err != nil {
			if ctx.Err() != nil {
				slog.Info("checker: result writer stopping")
				return
			}
			slog.Error("checker: read result", "err", err)
			continue
		}

		saveCtx, saveCancel := context.WithTimeout(context.Background(), 5*time.Second)

		check := &monitor.Check{
			MonitorID:  result.MonitorID,
			StatusCode: result.StatusCode,
			ResponseMs: result.ResponseMs,
		}
		if result.Error != nil {
			check.Error = result.Error
		}

		if err := c.repo.SaveCheck(saveCtx, check); err != nil {
			slog.Error("checker: save check to DB", "err", err)
		}

		m, getErr := c.repo.GetByID(saveCtx, result.MonitorID)
		if getErr == nil {
			now := time.Now()
			m.LastCheckedAt = &now
			if err := c.repo.Update(saveCtx, m); err != nil {
				slog.Error("checker: update last_checked_at", "err", err)
			}
		}

		saveCancel()

		if err := consumer.Commit(ctx, msg); err != nil {
			slog.Error("checker: commit result offset", "err", err)
		}
	}
}

func doProbe(ctx context.Context, url string) (int, string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Sprintf("request error: %s", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()
	return resp.StatusCode, ""
}
