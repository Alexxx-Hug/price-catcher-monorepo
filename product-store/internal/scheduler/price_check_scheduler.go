package scheduler

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type PriceCheckPublisher interface {
	PublishPriceCheckTasks(ctx context.Context, limit int) error
}

type PriceCheckScheduler struct {
	publisher PriceCheckPublisher
	interval  time.Duration
	limit     int
	logger    *zap.Logger
}

func NewPriceCheckScheduler(
	publisher PriceCheckPublisher,
	interval time.Duration,
	limit int,
	logger *zap.Logger,
) *PriceCheckScheduler {
	if logger == nil {
		logger = zap.NewNop()
	}

	if interval <= 0 {
		interval = 5 * time.Minute
	}

	if limit <= 0 {
		limit = 5
	}

	return &PriceCheckScheduler{
		publisher: publisher,
		interval:  interval,
		limit:     limit,
		logger:    logger,
	}
}

func (s *PriceCheckScheduler) Run(ctx context.Context) error {
	if err := s.publish(ctx); err != nil {
		s.logger.Error("failed to publish initial price check tasks", zap.Error(err))
	}

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := s.publish(ctx); err != nil {
				s.logger.Error("failed to publish price check tasks", zap.Error(err))
			}
		}
	}
}

func (s *PriceCheckScheduler) publish(ctx context.Context) error {
	if err := s.publisher.PublishPriceCheckTasks(ctx, s.limit); err != nil {
		return err
	}

	s.logger.Info("price check tasks published", zap.Int("limit", s.limit))
	return nil
}
