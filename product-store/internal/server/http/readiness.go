package http

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

type CheckFunc func(ctx context.Context) error

type ReadinessChecker struct {
	check    CheckFunc
	interval time.Duration
	timeout  time.Duration
	logger   *zap.Logger

	mu      sync.RWMutex
	ready   bool
	lastErr string
}

func NewReadinessChecker(check CheckFunc, interval time.Duration, timeout time.Duration, logger *zap.Logger) *ReadinessChecker {
	if logger == nil {
		logger = zap.NewNop()
	}

	if interval <= 0 {
		interval = 10 * time.Second
	}

	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	return &ReadinessChecker{
		check:    check,
		interval: interval,
		timeout:  timeout,
		logger:   logger,
	}
}

func (c *ReadinessChecker) Run(ctx context.Context) error {
	c.refresh(ctx)

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			c.refresh(ctx)
		}
	}
}

func (c *ReadinessChecker) IsReady() (bool, string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.ready, c.lastErr
}

func (c *ReadinessChecker) refresh(ctx context.Context) {
	checkCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	err := c.check(checkCtx)
	c.mu.Lock()
	defer c.mu.Unlock()

	if err != nil {
		c.ready = false
		c.lastErr = err.Error()
		c.logger.Error("readiness check failed", zap.Error(err))
		return
	}

	c.ready = true
	c.lastErr = ""
}
