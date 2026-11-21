package bootstrap

import (
	"errors"
	"fmt"
	"os"
	"time"

	"simple-server-template/app/platform/telemetry"
)

const serviceName = "simple-server-template"

// Version is set at build time:
//
//	go build -ldflags "-X simple-server-template/app/bootstrap.Version=1.2.3"
var Version = "dev"

type Config struct {
	HTTP      HTTPConfig
	PostsAPI  RemoteAPIConfig
	Telemetry telemetry.Config
}

type HTTPConfig struct {
	Addr      string
	AdminAddr string
}

type RemoteAPIConfig struct {
	BaseURL string
	Timeout time.Duration
	Retries int
}

const (
	envHTTPAddr  = "HTTP_ADDR"
	envAdminAddr = "ADMIN_ADDR"

	envPostsAPIURL     = "POSTS_API_URL"
	envPostsAPITimeout = "POSTS_API_TIMEOUT"

	// standard OpenTelemetry variable: prometheus | otlp | none
	envMetricsExporter = "OTEL_METRICS_EXPORTER"
)

func LoadConfig() (Config, error) {
	cfg := Config{
		HTTP: HTTPConfig{
			Addr:      envOr(envHTTPAddr, ":8080"),
			AdminAddr: envOr(envAdminAddr, ":9090"),
		},
		PostsAPI: RemoteAPIConfig{
			BaseURL: os.Getenv(envPostsAPIURL),
			Timeout: 5 * time.Second,
			Retries: 2,
		},
		Telemetry: telemetry.Config{
			ServiceName:    serviceName,
			ServiceVersion: Version,
			Exporter:       telemetry.ExporterPrometheus,
		},
	}

	if v := os.Getenv(envMetricsExporter); v != "" {
		cfg.Telemetry.Exporter = telemetry.Exporter(v)
	}

	if cfg.PostsAPI.BaseURL == "" {
		return Config{}, errors.New(envPostsAPIURL + " is not set")
	}

	if v := os.Getenv(envPostsAPITimeout); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("%s: %w", envPostsAPITimeout, err)
		}
		cfg.PostsAPI.Timeout = d
	}

	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
