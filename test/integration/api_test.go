// Package integration provides integration tests for the API.
package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"url-shorterner/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testRouter *gin.Engine
	testCfg    *config.Config
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	// Load test configuration
	cfg, err := SetupTestConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to setup test config: %v", err))
	}
	testCfg = cfg

	// Setup test router
	testRouter = SetupTestRouter(cfg)

	// Run tests
	code := m.Run()

	// Cleanup if needed
	os.Exit(code)
}

func TestShortenURL(t *testing.T) {
	reqBody := map[string]interface{}{
		"url": "https://example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Contains(t, resp, "short_code")
	assert.Contains(t, resp, "short_url")
	assert.Contains(t, resp, "expires_at")
	assert.NotEmpty(t, resp["short_code"])
}

func TestShortenURLWithAlias(t *testing.T) {
	alias := fmt.Sprintf("test-alias-%d", time.Now().Unix())
	reqBody := map[string]interface{}{
		"url":   "https://example.com",
		"alias": alias,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, alias, resp["short_code"])
}

func TestShortenURLWithExpiration(t *testing.T) {
	expiresIn := 3600
	reqBody := map[string]interface{}{
		"url":        "https://example.com",
		"expires_in": expiresIn,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.NotNil(t, resp["expires_at"])
}

func TestShortenURLInvalidURL(t *testing.T) {
	reqBody := map[string]interface{}{
		"url": "invalid-url",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp, "error")
}

func TestShortenURLDuplicateAlias(t *testing.T) {
	alias := fmt.Sprintf("duplicate-%d", time.Now().Unix())

	// First request
	reqBody1 := map[string]interface{}{
		"url":   "https://example.com",
		"alias": alias,
	}
	body1, _ := json.Marshal(reqBody1)

	req1 := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(body1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	testRouter.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request with same alias
	reqBody2 := map[string]interface{}{
		"url":   "https://google.com",
		"alias": alias,
	}
	body2, _ := json.Marshal(reqBody2)

	req2 := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	testRouter.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusConflict, w2.Code)
}

func TestShortenBatch(t *testing.T) {
	reqBody := map[string]interface{}{
		"items": []map[string]interface{}{
			{"url": "https://example.com"},
			{"url": "https://google.com", "alias": fmt.Sprintf("batch-%d", time.Now().Unix())},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/shorten/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Contains(t, resp, "results")
	results := resp["results"].([]interface{})
	assert.Len(t, results, 2)
}

func TestRedirect(t *testing.T) {
	// First, create a shortened URL
	reqBody := map[string]interface{}{
		"url": "https://example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	var shortenResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &shortenResp)
	require.NoError(t, err)
	shortCode := shortenResp["short_code"].(string)

	// Now test redirect
	req = httptest.NewRequest(http.MethodGet, "/"+shortCode, nil)
	w = httptest.NewRecorder()

	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Location"))
}

func TestRedirectNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/nonexistent-code-12345", nil)
	w := httptest.NewRecorder()

	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetAnalytics(t *testing.T) {
	// First, create a shortened URL
	reqBody := map[string]interface{}{
		"url": "https://example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	var shortenResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &shortenResp)
	require.NoError(t, err)
	shortCode := shortenResp["short_code"].(string)

	// Get analytics
	req = httptest.NewRequest(http.MethodGet, "/analytics/"+shortCode, nil)
	w = httptest.NewRecorder()

	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var analyticsResp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &analyticsResp)
	require.NoError(t, err)

	assert.Equal(t, shortCode, analyticsResp["short_code"])
	assert.Contains(t, analyticsResp, "total_clicks")
	assert.Contains(t, analyticsResp, "unique_ips")
	assert.Contains(t, analyticsResp, "records")
}

func TestGetAnalyticsWithLimit(t *testing.T) {
	// First, create a shortened URL
	reqBody := map[string]interface{}{
		"url": "https://example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	var shortenResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &shortenResp)
	require.NoError(t, err)
	shortCode := shortenResp["short_code"].(string)

	// Get analytics with limit
	req = httptest.NewRequest(http.MethodGet, "/analytics/"+shortCode+"?limit=50", nil)
	w = httptest.NewRecorder()

	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetAnalyticsNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/analytics/nonexistent-code", nil)
	w := httptest.NewRecorder()

	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code) // Returns empty analytics, not 404

	var analyticsResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &analyticsResp)
	require.NoError(t, err)
	assert.Equal(t, 0, int(analyticsResp["total_clicks"].(float64)))
}
