package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/henok/transit-backend/internal/delivery/http/handler"
	"github.com/henok/transit-backend/internal/delivery/http/middleware"
	"github.com/henok/transit-backend/internal/repository/postgres"
	"github.com/henok/transit-backend/internal/repository/redis"
	"github.com/rs/zerolog"
)

// NewRouter builds a gin engine with global middleware and registers
// all route groups. returns an http.Handler for use with the standard
// http.Server.
func NewRouter(pg *postgres.Pool, rdb *redis.Client, logger zerolog.Logger) http.Handler {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logging(logger))
	r.Use(middleware.Recover(logger))

	healthHandler := handler.NewHealthHandler(pg, rdb)

	r.GET("/healthz", healthHandler.Health)

	return r
}
