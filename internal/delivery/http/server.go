package http

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type Server struct {
	*http.Server
	logger zerolog.Logger
}

// NewServer wraps an http.Server with sensible timeouts and a logger
// for structured startup/shutdown messages.
func NewServer(addr string, handler http.Handler, logger zerolog.Logger) *Server {
	return &Server{
		Server: &http.Server{
			Addr:         ":" + addr,
			Handler:      handler,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		logger: logger,
	}
}

func (s *Server) Start() error {
	s.logger.Info().Str("addr", s.Addr).Msg("starting http server")
	return s.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info().Msg("shutting down http server")
	return s.Server.Shutdown(ctx)
}
