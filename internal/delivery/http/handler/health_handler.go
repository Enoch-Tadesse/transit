package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/henok/transit-backend/internal/delivery/http/response"
	"github.com/henok/transit-backend/internal/repository/postgres"
	"github.com/henok/transit-backend/internal/repository/redis"
)

type HealthHandler struct {
	pg  *postgres.Pool
	rdb *redis.Client
}

func NewHealthHandler(pg *postgres.Pool, rdb *redis.Client) *HealthHandler {
	return &HealthHandler{pg: pg, rdb: rdb}
}

// Health checks postgres and redis connectivity and returns a 200 or 503
// with per-service status so load balancers and orchestration tools know
// whether the instance is ready to serve traffic.
func (h *HealthHandler) Health(c *gin.Context) {
	// dont let a stuck dependency hang the health endpoint forever
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	services := make([]response.HealthzStatus, 0, 2)
	overallOK := true

	if err := h.pg.Health(ctx); err != nil {
		overallOK = false
		services = append(services, response.HealthzStatus{
			Service: "postgres",
			Status:  "unhealthy",
			Error:   err.Error(),
		})
	} else {
		services = append(services, response.HealthzStatus{
			Service: "postgres",
			Status:  "healthy",
		})
	}

	if err := h.rdb.Health(ctx); err != nil {
		overallOK = false
		services = append(services, response.HealthzStatus{
			Service: "redis",
			Status:  "unhealthy",
			Error:   err.Error(),
		})
	} else {
		services = append(services, response.HealthzStatus{
			Service: "redis",
			Status:  "healthy",
		})
	}

	if !overallOK {
		response.WriteHealthz(c, http.StatusServiceUnavailable, services)
		return
	}

	response.WriteHealthz(c, http.StatusOK, services)
}
