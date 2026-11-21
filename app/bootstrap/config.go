package bootstrap

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// Config - external settings of the application instance.
// Production values come from the environment (see LoadConfig),
// tests build it directly, e.g. with the mock server URL.
type Config struct {
	PostsAPI RemoteAPIConfig
}

type RemoteAPIConfig struct {
	BaseURL string
	Timeout time.Duration
	Retries int
}

const (
	envPostsAPIURL     = "POSTS_API_URL"
	envPostsAPITimeout = "POSTS_API_TIMEOUT"
)

// LoadConfig reads Config from environment variables.
//
//	POSTS_API_URL      required, base URL of the posts API
//	POSTS_API_TIMEOUT  optional, Go duration (default 5s)
func LoadConfig() (Config, error) {
	cfg := Config{
		PostsAPI: RemoteAPIConfig{
			BaseURL: os.Getenv(envPostsAPIURL),
			Timeout: 5 * time.Second,
			Retries: 2,
		},
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
