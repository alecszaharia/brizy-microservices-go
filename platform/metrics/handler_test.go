package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetricsHandler(t *testing.T) {
	reg := NewRegistry("test_service", "1.0.0")
	handler := NewMetricsHandler(reg)

	require.NotNil(t, handler)

	// Create test request
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	// Serve the request
	handler.ServeHTTP(w, req)

	// Check response
	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	assert.True(t, strings.Contains(contentType, "text/plain"), "Content-Type should be text/plain")

	// Read body
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	bodyStr := string(body)

	// Should contain Prometheus format markers
	assert.Contains(t, bodyStr, "# HELP", "Should contain HELP comments")
	assert.Contains(t, bodyStr, "# TYPE", "Should contain TYPE comments")

	// Should contain Go runtime metrics
	assert.Contains(t, bodyStr, "go_goroutines", "Should contain Go runtime metrics")

	// Should contain build info
	assert.Contains(t, bodyStr, "test_service_build_info", "Should contain build info metric")
}

func TestNewMetricsHandler_WithCustomMetrics(t *testing.T) {
	reg := NewRegistry("test_service", "1.0.0")

	// Add custom metrics
	counter := reg.NewCounter("custom_counter_total", "Custom counter")
	counter.Inc()

	gauge := reg.NewGauge("custom_gauge", "Custom gauge")
	gauge.Set(42)

	handler := NewMetricsHandler(reg)

	// Create test request
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	// Serve the request
	handler.ServeHTTP(w, req)

	// Read body
	body, err := io.ReadAll(w.Result().Body)
	require.NoError(t, err)
	bodyStr := string(body)

	// Should contain custom metrics
	assert.Contains(t, bodyStr, "test_service_custom_counter_total", "Should contain custom counter")
	assert.Contains(t, bodyStr, "test_service_custom_gauge", "Should contain custom gauge")
}

func TestNewMetricsHandler_NilRegistry(t *testing.T) {
	handler := NewMetricsHandler(nil)

	require.NotNil(t, handler)

	// Create test request
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	// Serve the request
	handler.ServeHTTP(w, req)

	// Check response
	resp := w.Result()
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "Metrics disabled")
}

func TestNewMetricsHandler_PrometheusFormat(t *testing.T) {
	reg := NewRegistry("test_service", "1.0.0")

	// Add a simple counter
	counter := reg.NewCounterVec("requests_total", "Total requests", []string{"method"})
	counter.WithLabelValues("GET").Add(5)
	counter.WithLabelValues("POST").Add(3)

	handler := NewMetricsHandler(reg)

	// Create test request
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	// Serve the request
	handler.ServeHTTP(w, req)

	body, err := io.ReadAll(w.Result().Body)
	require.NoError(t, err)
	bodyStr := string(body)

	// Verify Prometheus format
	assert.Contains(t, bodyStr, "# HELP test_service_requests_total Total requests")
	assert.Contains(t, bodyStr, "# TYPE test_service_requests_total counter")
	assert.Contains(t, bodyStr, `test_service_requests_total{method="GET"} 5`)
	assert.Contains(t, bodyStr, `test_service_requests_total{method="POST"} 3`)
}

func TestNewMetricsHandler_BuildInfo(t *testing.T) {
	reg := NewRegistry("my_service", "1.0.0")
	handler := NewMetricsHandler(reg)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	body, err := io.ReadAll(w.Result().Body)
	require.NoError(t, err)
	bodyStr := string(body)

	// Check for build info metric
	assert.Contains(t, bodyStr, "my_service_build_info")
	assert.Contains(t, bodyStr, `version="1.0.0"`)
	assert.Contains(t, bodyStr, `my_service_build_info{version="1.0.0"} 1`)
}
