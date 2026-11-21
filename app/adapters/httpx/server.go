package httpx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

type Server struct {
	Name string
	*http.Server
	listener net.Listener
}

func Listen(name, addr string, handler http.Handler) (*Server, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("%s server: %w", name, err)
	}
	return NewServer(name, ln, handler), nil
}

// NewServer wraps an already open listener.
func NewServer(name string, ln net.Listener, handler http.Handler) *Server {
	return &Server{
		Name: name,
		Server: &http.Server{
			Handler:           handler,
			ReadHeaderTimeout: 10 * time.Second,
		},
		listener: ln,
	}
}

func (s *Server) Addr() string {
	return s.listener.Addr().String()
}

func Serve(ctx context.Context, log *slog.Logger, shutdownTimeout time.Duration, servers ...*Server) error {
	failed := make(chan error, len(servers))
	for _, s := range servers {
		log.Info("http server started", "server", s.Name, "addr", s.Addr())
		go func() {
			if err := s.Serve(s.listener); !errors.Is(err, http.ErrServerClosed) {
				failed <- fmt.Errorf("%s server: %w", s.Name, err)
			}
		}()
	}

	var runErr error
	select {
	case <-ctx.Done():
		log.Info("shutting down http servers")
	case runErr = <-failed:
		log.Error("http server failed, shutting down", "error", runErr)
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()

	errs := []error{runErr}
	for _, s := range servers {
		if err := s.Shutdown(shutdownCtx); err != nil {
			errs = append(errs, fmt.Errorf("%s server shutdown: %w", s.Name, err))
		}
	}
	return errors.Join(errs...)
}
