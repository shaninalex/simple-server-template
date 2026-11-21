package telemetry

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	promclient "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

type Exporter string

const (
	ExporterNone       Exporter = "none"
	ExporterPrometheus Exporter = "prometheus"
	ExporterOTLP       Exporter = "otlp"
)

type Config struct {
	ServiceName    string
	ServiceVersion string
	Exporter       Exporter
	ExportInterval time.Duration
}

type Telemetry struct {
	provider *sdkmetric.MeterProvider
	metrics  http.Handler
}

func New(ctx context.Context, cfg Config) (*Telemetry, error) {
	res, err := newResource(ctx, cfg)
	if err != nil {
		return nil, err
	}

	t := &Telemetry{metrics: http.NotFoundHandler()}
	opts := []sdkmetric.Option{sdkmetric.WithResource(res)}

	switch cfg.Exporter {
	case ExporterNone, "":
	case ExporterPrometheus:
		registry := promclient.NewRegistry()
		exp, err := prometheus.New(prometheus.WithRegisterer(registry))
		if err != nil {
			return nil, fmt.Errorf("telemetry: prometheus exporter: %w", err)
		}
		opts = append(opts, sdkmetric.WithReader(exp))
		t.metrics = promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	case ExporterOTLP:
		exp, err := otlpmetrichttp.New(ctx)
		if err != nil {
			return nil, fmt.Errorf("telemetry: otlp exporter: %w", err)
		}
		interval := cfg.ExportInterval
		if interval <= 0 {
			interval = 30 * time.Second
		}
		opts = append(opts, sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(exp, sdkmetric.WithInterval(interval)),
		))
	default:
		return nil, fmt.Errorf("telemetry: unknown exporter %q", cfg.Exporter)
	}

	t.provider = sdkmetric.NewMeterProvider(opts...)
	if err := runtime.Start(runtime.WithMeterProvider(t.provider)); err != nil {
		_ = t.provider.Shutdown(ctx)
		return nil, fmt.Errorf("telemetry: runtime metrics: %w", err)
	}

	return t, nil
}

func newResource(ctx context.Context, cfg Config) (*resource.Resource, error) {
	var attrs []resource.Option
	if cfg.ServiceName != "" {
		attrs = append(attrs, resource.WithAttributes(semconv.ServiceName(cfg.ServiceName)))
	}
	if cfg.ServiceVersion != "" {
		attrs = append(attrs, resource.WithAttributes(semconv.ServiceVersion(cfg.ServiceVersion)))
	}

	res, err := resource.New(ctx, append(attrs,
		resource.WithTelemetrySDK(),
		resource.WithProcessRuntimeName(),
		resource.WithProcessRuntimeVersion(),
		resource.WithFromEnv(),
	)...)
	// A partial resource with a schema URL conflict is still usable.
	if err != nil && !errors.Is(err, resource.ErrSchemaURLConflict) {
		return nil, fmt.Errorf("telemetry: resource: %w", err)
	}
	return res, nil
}

func (t *Telemetry) MeterProvider() metric.MeterProvider {
	return t.provider
}

func (t *Telemetry) MetricsHandler() http.Handler {
	return t.metrics
}

func (t *Telemetry) AdminHandler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", t.metrics)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return mux
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	return t.provider.Shutdown(ctx)
}
