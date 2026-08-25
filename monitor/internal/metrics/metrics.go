package metrics

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type Metrics struct {
	KafkaMessagesConsumed *prometheus.CounterVec
	KafkaMessagesProduced *prometheus.CounterVec
	KafkaErrors           *prometheus.CounterVec
	KafkaProcessingTime   *prometheus.HistogramVec
}

func New(serviceName string) (*Metrics, *prometheus.Registry) {
	registry := prometheus.NewRegistry()
	registry.MustRegister(prometheus.NewGoCollector(), prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	constLabels := prometheus.Labels{"service": serviceName}
	m := &Metrics{
		KafkaMessagesConsumed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        "kafka_messages_consumed_total",
			Help:        "Total number of Kafka messages consumed.",
			ConstLabels: constLabels,
		}, []string{"topic", "status"}),
		KafkaMessagesProduced: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        "kafka_messages_produced_total",
			Help:        "Total number of Kafka messages produced.",
			ConstLabels: constLabels,
		}, []string{"topic", "status"}),
		KafkaErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        "kafka_errors_total",
			Help:        "Total number of Kafka processing errors.",
			ConstLabels: constLabels,
		}, []string{"topic", "operation"}),
		KafkaProcessingTime: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:        "kafka_message_processing_seconds",
			Help:        "Kafka message processing duration in seconds.",
			ConstLabels: constLabels,
			Buckets:     prometheus.DefBuckets,
		}, []string{"topic", "status"}),
	}

	registry.MustRegister(m.KafkaMessagesConsumed, m.KafkaMessagesProduced, m.KafkaErrors, m.KafkaProcessingTime)
	return m, registry
}

func (m *Metrics) ObserveConsumed(topic string, status string, startedAt time.Time) {
	if m == nil {
		return
	}

	m.KafkaMessagesConsumed.WithLabelValues(topic, status).Inc()
	m.KafkaProcessingTime.WithLabelValues(topic, status).Observe(time.Since(startedAt).Seconds())
}

func (m *Metrics) IncProduced(topic string, err error) {
	if m == nil {
		return
	}

	status := "success"
	if err != nil {
		status = "error"
		m.KafkaErrors.WithLabelValues(topic, "produce").Inc()
	}

	m.KafkaMessagesProduced.WithLabelValues(topic, status).Inc()
}

func (m *Metrics) IncError(topic string, operation string) {
	if m == nil {
		return
	}

	m.KafkaErrors.WithLabelValues(topic, operation).Inc()
}

func RunServer(ctx context.Context, port string, registry *prometheus.Registry, logger *zap.Logger) error {
	if port == "" {
		return nil
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	errChan := make(chan error, 1)
	go func() {
		logger.Info("metrics server started", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown metrics server: %w", err)
		}

		return nil
	case err := <-errChan:
		return err
	}
}
