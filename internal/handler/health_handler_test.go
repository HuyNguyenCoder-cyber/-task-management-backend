package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type fakeSQLPinger struct {
	err   error
	delay time.Duration
}

func (f fakeSQLPinger) PingContext(ctx context.Context) error {
	if f.delay > 0 {
		timer := time.NewTimer(f.delay)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	}

	return f.err
}

type fakeRedisPinger struct {
	err   error
	delay time.Duration
}

func (f fakeRedisPinger) Ping(ctx context.Context) error {
	if f.delay > 0 {
		timer := time.NewTimer(f.delay)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	}

	return f.err
}

func TestHealthHandlerCheck_UP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHealthHandler(
		fakeSQLPinger{},
		fakeRedisPinger{},
	)

	router := gin.New()
	router.GET("/health", handler.Check)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var payload HealthResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &payload)
	assert.NoError(t, err)
	assert.Equal(t, "UP", payload.Status)
	assert.Equal(t, "UP", payload.Services["database"].Status)
	assert.Equal(t, "UP", payload.Services["redis"].Status)
	assert.NotEmpty(t, payload.Services["database"].Details)
	assert.NotEmpty(t, payload.Services["redis"].Details)
}

func TestHealthHandlerCheck_DownWhenDatabaseFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHealthHandler(
		fakeSQLPinger{err: errors.New("mysql unavailable")},
		fakeRedisPinger{},
	)

	router := gin.New()
	router.GET("/health", handler.Check)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	var payload HealthResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &payload)
	assert.NoError(t, err)
	assert.Equal(t, "DOWN", payload.Status)
	assert.Equal(t, "DOWN", payload.Services["database"].Status)
	assert.Contains(t, payload.Services["database"].Error, "mysql unavailable")
}

func TestHealthHandlerCheck_TimesOut(t *testing.T) {
	gin.SetMode(gin.TestMode)

	oldTimeout := healthCheckTimeout
	healthCheckTimeout = 10 * time.Millisecond
	t.Cleanup(func() {
		healthCheckTimeout = oldTimeout
	})

	handler := NewHealthHandler(
		fakeSQLPinger{delay: 200 * time.Millisecond},
		fakeRedisPinger{},
	)

	router := gin.New()
	router.GET("/health", handler.Check)

	start := time.Now()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(recorder, request)
	elapsed := time.Since(start)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Less(t, elapsed, 100*time.Millisecond)

	var payload HealthResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &payload)
	assert.NoError(t, err)
	assert.Equal(t, "DOWN", payload.Status)
	assert.Equal(t, "DOWN", payload.Services["database"].Status)
	assert.Contains(t, payload.Services["database"].Error, "context deadline exceeded")
}
