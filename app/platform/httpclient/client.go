package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"simple-server-template/app/core/errs"
)

type Config struct {
	BaseURL      string
	Timeout      time.Duration
	Retries      int
	RetryBackoff time.Duration
	MaxBodyBytes int64
	UserAgent    string
}

type Client struct {
	base *url.URL
	http *http.Client
	cfg  Config
}

func New(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, errors.New("httpclient: empty BaseURL")
	}
	base, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("httpclient: parse BaseURL: %w", err)
	}
	if base.Scheme != "http" && base.Scheme != "https" {
		return nil, fmt.Errorf("httpclient: BaseURL must be http(s), got %q", cfg.BaseURL)
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Second
	}
	if cfg.RetryBackoff <= 0 {
		cfg.RetryBackoff = 200 * time.Millisecond
	}
	if cfg.MaxBodyBytes <= 0 {
		cfg.MaxBodyBytes = 4 << 20
	}
	return &Client{
		base: base,
		http: &http.Client{Timeout: cfg.Timeout},
		cfg:  cfg,
	}, nil
}

func (c *Client) GetJSON(ctx context.Context, path string, query url.Values, out any) error {
	u := c.base.JoinPath(path)
	u.RawQuery = query.Encode()
	target := u.String()

	var lastErr error
	backoff := c.cfg.RetryBackoff
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return errs.Platform("HTTPC00", fmt.Sprintf("GET %s: %v (last error: %v)", target, ctx.Err(), lastErr))
			case <-time.After(backoff):
			}
			backoff *= 2
		}

		retry, err := c.getOnce(ctx, target, out)
		if err == nil {
			return nil
		}
		lastErr = err
		if !retry {
			break
		}
	}
	return lastErr
}

func (c *Client) getOnce(ctx context.Context, target string, out any) (retry bool, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return false, errs.Platform("HTTPC01", fmt.Sprintf("GET %s: build request: %v", target, err))
	}
	req.Header.Set("Accept", "application/json")
	if c.cfg.UserAgent != "" {
		req.Header.Set("User-Agent", c.cfg.UserAgent)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		retry = ctx.Err() == nil
		return retry, errs.Platform("HTTPC02", fmt.Sprintf("GET %s: %v", target, err))
	}
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, c.cfg.MaxBodyBytes))
		_ = resp.Body.Close()
	}()

	switch code := resp.StatusCode; {
	case code >= 200 && code < 300:
		if err := json.NewDecoder(io.LimitReader(resp.Body, c.cfg.MaxBodyBytes)).Decode(out); err != nil {
			return false, errs.Platform("HTTPC03", fmt.Sprintf("GET %s: decode response: %v", target, err))
		}
		return false, nil
	case code == http.StatusNotFound:
		return false, errs.NotFound("HTTPC04", fmt.Sprintf("GET %s: not found", target))
	case code == http.StatusTooManyRequests:
		return true, errs.RateLimited("HTTPC05", fmt.Sprintf("GET %s: rate limited", target))
	case code >= 500:
		return true, errs.Platform("HTTPC06", fmt.Sprintf("GET %s: upstream status %d", target, code))
	default:
		return false, errs.Platform("HTTPC07", fmt.Sprintf("GET %s: unexpected status %d", target, code))
	}
}
