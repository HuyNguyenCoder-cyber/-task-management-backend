package handler

import (
	"context"
	"database/sql"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var healthCheckTimeout = 2 * time.Second

type SQLPinger interface {
	PingContext(context.Context) error
}

type RedisPinger interface {
	Ping(context.Context) error
}

type HealthHandler struct {
	sqlDB SQLPinger
	redis RedisPinger
}

type HealthResponse struct {
	Status   string                   `json:"status"`
	Services map[string]HealthService `json:"services"`
}

type HealthService struct {
	Status  string `json:"status"`
	Details string `json:"details,omitempty"`
	Error   string `json:"error,omitempty"`
}

type healthCheckResult struct {
	name   string
	health HealthService
}

func NewHealthHandler(sqlDB SQLPinger, redis RedisPinger) *HealthHandler {
	return &HealthHandler{
		sqlDB: sqlDB,
		redis: redis,
	}
}

func NewRedisPinger(client *redis.Client) RedisPinger {
	return redisClientPinger{client: client}
}

func (h *HealthHandler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), healthCheckTimeout)
	defer cancel()

	results := make(chan healthCheckResult, 2)
	var once sync.Once

	go func() {
		results <- healthCheckResult{
			name:   "database",
			health: pingSQLDB(ctx, h.sqlDB),
		}
	}()

	go func() {
		results <- healthCheckResult{
			name:   "redis",
			health: pingRedis(ctx, h.redis),
		}
	}()

	services := map[string]HealthService{
		"database": {},
		"redis":    {},
	}

	pending := 2
	for pending > 0 {
		select {
		case result := <-results:
			services[result.name] = result.health
			pending--
		case <-ctx.Done():
			once.Do(func() {
				timeoutHealth := HealthService{
					Status: "DOWN",
					Error:  ctx.Err().Error(),
				}
				if services["database"].Status == "" {
					services["database"] = timeoutHealth
				}
				if services["redis"].Status == "" {
					services["redis"] = timeoutHealth
				}
			})
			pending = 0
		}
	}

	status := "UP"
	httpStatus := http.StatusOK
	if services["database"].Status != "UP" || services["redis"].Status != "UP" {
		status = "DOWN"
		httpStatus = http.StatusInternalServerError
	}

	c.JSON(httpStatus, HealthResponse{
		Status:   status,
		Services: services,
	})
}

func pingSQLDB(ctx context.Context, pinger SQLPinger) HealthService {
	if err := pinger.PingContext(ctx); err != nil {
		return HealthService{
			Status: "DOWN",
			Error:  err.Error(),
		}
	}

	return HealthService{
		Status:  "UP",
		Details: "database connection is healthy",
	}
}

func pingRedis(ctx context.Context, pinger RedisPinger) HealthService {
	if err := pinger.Ping(ctx); err != nil {
		return HealthService{
			Status: "DOWN",
			Error:  err.Error(),
		}
	}

	return HealthService{
		Status:  "UP",
		Details: "redis connection is healthy",
	}
}

type redisClientPinger struct {
	client *redis.Client
}

func (r redisClientPinger) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

var _ SQLPinger = (*sql.DB)(nil)
