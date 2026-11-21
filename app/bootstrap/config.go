package bootstrap

import (
	"errors"
	"fmt"
	"os"
	"time"
)

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
