package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func registerAndLogin(t *testing.T, router *gin.Engine, email, password, fullName string) (int64, string) {
	t.Helper()

	regBody := map[string]any{
		"email":     email,
		"password":  password,
		"full_name": fullName,
	}
	regRR := performRequest(router, http.MethodPost, "/auth/register", regBody, "")
	assert.Equal(t, http.StatusCreated, regRR.Code)

	var regResp apiResponse
	err := json.Unmarshal(regRR.Body.Bytes(), &regResp)
	assert.NoError(t, err)
	assert.True(t, regResp.Success)

	var regData registerData
	err = json.Unmarshal(regResp.Data, &regData)
	assert.NoError(t, err)
	assert.Greater(t, regData.ID, int64(0))

	loginBody := map[string]any{
		"email":    email,
		"password": password,
	}
	loginRR := performRequest(router, http.MethodPost, "/auth/login", loginBody, "")
	assert.Equal(t, http.StatusOK, loginRR.Code)

	var loginResp apiResponse
	err = json.Unmarshal(loginRR.Body.Bytes(), &loginResp)
	assert.NoError(t, err)
	assert.True(t, loginResp.Success)

	var loginPayload loginData
	err = json.Unmarshal(loginResp.Data, &loginPayload)
	assert.NoError(t, err)
	assert.NotEmpty(t, loginPayload.AccessToken)

	return regData.ID, loginPayload.AccessToken
}

func TestSecurityIntegration_AuthMiddlewareAndProjectAuthorization(t *testing.T) {
	router, db, redisClient := setupE2ERouter(t)

	var (
		userAID  int64
		userBID  int64
		projectID int64
	)

	emailA := fmt.Sprintf("sec_a_%d@example.com", time.Now().UnixNano())
	emailB := fmt.Sprintf("sec_b_%d@example.com", time.Now().UnixNano()+1)
	password := "123456"

	t.Cleanup(func() {
		ctx := context.Background()
		if projectID > 0 {
			db.Exec("DELETE FROM project_members WHERE project_id = ?", projectID)
			db.Exec("DELETE FROM projects WHERE id = ?", projectID)
		}
		if userAID > 0 {
			db.Exec("DELETE FROM users WHERE id = ?", userAID)
		} else {
			db.Exec("DELETE FROM users WHERE email = ?", emailA)
		}
		if userBID > 0 {
			db.Exec("DELETE FROM users WHERE id = ?", userBID)
		} else {
			db.Exec("DELETE FROM users WHERE email = ?", emailB)
		}
		if redisClient != nil {
			_ = redisClient.Del(ctx, "queue:notifications").Err()
			_ = redisClient.Close()
		}
	})

	t.Run("Task 1 - Authentication Middleware", func(t *testing.T) {
		body := map[string]any{
			"name":        "Protected Project",
			"description": "should fail without valid token",
		}

		t.Run("1. No Token returns 401", func(t *testing.T) {
			rr := performRequest(router, http.MethodPost, "/projects", body, "")
			assert.Equal(t, http.StatusUnauthorized, rr.Code)
		})

		t.Run("2. Invalid Token returns 401", func(t *testing.T) {
			rr := performRequest(router, http.MethodPost, "/projects", body, "invalid_token_xyz")
			assert.Equal(t, http.StatusUnauthorized, rr.Code)
		})
	})

	t.Run("Task 2 - Business Rule Authorization", func(t *testing.T) {
		var tokenA, tokenB string

		t.Run("1. User A register/login and create project", func(t *testing.T) {
			userAID, tokenA = registerAndLogin(t, router, emailA, password, "Security User A")

			createBody := map[string]any{
				"name":        "Project of User A",
				"description": "authorization testing",
			}
			rr := performRequest(router, http.MethodPost, "/projects", createBody, tokenA)
			assert.Equal(t, http.StatusCreated, rr.Code)

			var resp apiResponse
			err := json.Unmarshal(rr.Body.Bytes(), &resp)
			assert.NoError(t, err)
			assert.True(t, resp.Success)

			var data projectData
			err = json.Unmarshal(resp.Data, &data)
			assert.NoError(t, err)
			assert.Greater(t, data.ID, int64(0))
			projectID = data.ID
		})

		t.Run("2. User B register/login", func(t *testing.T) {
			userBID, tokenB = registerAndLogin(t, router, emailB, password, "Security User B")
		})

		t.Run("3. User B cannot update/delete User A project", func(t *testing.T) {
			updatePath := fmt.Sprintf("/projects/%d", projectID)
			updateBody := map[string]any{
				"name":        "Hacked Name",
				"description": "should be rejected",
			}
			updateRR := performRequest(router, http.MethodPut, updatePath, updateBody, tokenB)
			assert.Contains(t, []int{http.StatusForbidden, http.StatusUnauthorized}, updateRR.Code)

			deletePath := fmt.Sprintf("/projects/%d", projectID)
			deleteRR := performRequest(router, http.MethodDelete, deletePath, nil, tokenB)
			assert.Contains(t, []int{http.StatusForbidden, http.StatusUnauthorized}, deleteRR.Code)
		})
	})
}
