package metrics

import (
	"platform/build"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegistry(t *testing.T) {
	reg := NewRegistry(build.NewBuildInfo("test_service", "1.0.0"))

	require.NotNil(t, reg)
	assert.NotNil(t, reg.registry)
	assert.Equal(t, "test_service", reg.serviceName)

	// Verify that Go collectors are registered
	metrics, err := reg.Unwrap().Gather()
	require.NoError(t, err)

	// Check for runtime metrics
	hasGoMetrics := false
	hasBuildInfo := false
	for _, m := range metrics {
		if m.GetName() == "go_goroutines" {
			hasGoMetrics = true
		}
		if m.GetName() == "test_service_build_info" {
			hasBuildInfo = true
		}
	}
	assert.True(t, hasGoMetrics, "Should have Go runtime metrics")
	assert.True(t, hasBuildInfo, "Should have build info metric")
}

func TestRegistry_NewCounterVec(t *testing.T) {
	reg := NewRegistry(build.NewBuildInfo("my_service", "1.0.0"))

	counter := reg.NewCounterVec("requests_total", "Total requests", []string{"method", "status"})
	require.NotNil(t, counter)

	// Increment counter
	counter.WithLabelValues("GET", "200").Inc()
	counter.WithLabelValues("POST", "201").Inc()

	// Gather metrics
	metrics, err := reg.Unwrap().Gather()
	require.NoError(t, err)

	// Find our counter
	found := false
	for _, m := range metrics {
		if m.GetName() == "my_service_requests_total" {
			found = true
			assert.Equal(t, "Total requests", m.GetHelp())
			assert.Equal(t, 2, len(m.GetMetric()))
		}
	}
	assert.True(t, found, "Counter should be registered with service name prefix")
}

func TestRegistry_NewGaugeVec(t *testing.T) {
	reg := NewRegistry(build.NewBuildInfo("my_service", "1.0.0"))

	gauge := reg.NewGaugeVec("queue_depth", "Current queue depth", []string{"queue"})
	require.NotNil(t, gauge)

	// Set gauge values
	gauge.WithLabelValues("high_priority").Set(10)
	gauge.WithLabelValues("low_priority").Set(5)

	// Gather metrics
	metrics, err := reg.Unwrap().Gather()
	require.NoError(t, err)

	// Find our gauge
	found := false
	for _, m := range metrics {
		if m.GetName() == "my_service_queue_depth" {
			found = true
			assert.Equal(t, "Current queue depth", m.GetHelp())
			assert.Equal(t, 2, len(m.GetMetric()))
		}
	}
	assert.True(t, found, "Gauge should be registered with service name prefix")
}

func TestRegistry_NewHistogramVec(t *testing.T) {
	reg := NewRegistry(build.NewBuildInfo("my_service", "1.0.0"))

	buckets := []float64{0.1, 0.5, 1.0}
	histogram := reg.NewHistogramVec("request_duration_seconds", "Request duration", buckets, []string{"endpoint"})
	require.NotNil(t, histogram)

	// Record observations
	histogram.WithLabelValues("/api/users").Observe(0.15)
	histogram.WithLabelValues("/api/users").Observe(0.75)
	histogram.WithLabelValues("/api/posts").Observe(0.05)

	// Gather metrics
	metrics, err := reg.Unwrap().Gather()
	require.NoError(t, err)

	// Find our histogram
	found := false
	for _, m := range metrics {
		if m.GetName() == "my_service_request_duration_seconds" {
			found = true
			assert.Equal(t, "Request duration", m.GetHelp())
		}
	}
	assert.True(t, found, "Histogram should be registered with service name prefix")
}

func TestRegistry_NewCounter(t *testing.T) {
	reg := NewRegistry(build.NewBuildInfo("my_service", "1.0.0"))

	counter := reg.NewCounter("operations_total", "Total operations")
	require.NotNil(t, counter)

	// Increment counter
	counter.Inc()
	counter.Add(5)

	// Gather metrics
	metrics, err := reg.Unwrap().Gather()
	require.NoError(t, err)

	// Find our counter
	found := false
	for _, m := range metrics {
		if m.GetName() == "my_service_operations_total" {
			found = true
			assert.Equal(t, "Total operations", m.GetHelp())
			assert.Equal(t, 1, len(m.GetMetric()))
			assert.Equal(t, float64(6), m.GetMetric()[0].GetCounter().GetValue())
		}
	}
	assert.True(t, found, "Counter should be registered with service name prefix")
}

func TestRegistry_NewGauge(t *testing.T) {
	reg := NewRegistry(build.NewBuildInfo("my_service", "1.0.0"))

	gauge := reg.NewGauge("temperature", "Current temperature")
	require.NotNil(t, gauge)

	// Set gauge value
	gauge.Set(23.5)

	// Gather metrics
	metrics, err := reg.Unwrap().Gather()
	require.NoError(t, err)

	// Find our gauge
	found := false
	for _, m := range metrics {
		if m.GetName() == "my_service_temperature" {
			found = true
			assert.Equal(t, "Current temperature", m.GetHelp())
			assert.Equal(t, 1, len(m.GetMetric()))
			assert.Equal(t, float64(23.5), m.GetMetric()[0].GetGauge().GetValue())
		}
	}
	assert.True(t, found, "Gauge should be registered with service name prefix")
}

func TestRegistry_MultipleRegistries_Isolated(t *testing.T) {
	// Create two registries for different services
	reg1 := NewRegistry(build.NewBuildInfo("service_one", "1.0.0"))
	reg2 := NewRegistry(build.NewBuildInfo("service_two", "1.0.0"))

	// Register same metric name in both
	counter1 := reg1.NewCounter("requests_total", "Service one requests")
	counter2 := reg2.NewCounter("requests_total", "Service two requests")

	counter1.Add(10)
	counter2.Add(20)

	// Check service one metrics
	metrics1, err := reg1.Unwrap().Gather()
	require.NoError(t, err)
	found := false
	for _, m := range metrics1 {
		if m.GetName() == "service_one_requests_total" {
			found = true
			assert.Equal(t, float64(10), m.GetMetric()[0].GetCounter().GetValue())
		}
	}
	assert.True(t, found, "Service one metric should exist")

	// Check service two metrics
	metrics2, err := reg2.Unwrap().Gather()
	require.NoError(t, err)
	found = false
	for _, m := range metrics2 {
		if m.GetName() == "service_two_requests_total" {
			found = true
			assert.Equal(t, float64(20), m.GetMetric()[0].GetCounter().GetValue())
		}
	}
	assert.True(t, found, "Service two metric should exist")
}

func TestRegistry_DuplicateMetric_Panics(t *testing.T) {
	reg := NewRegistry(build.NewBuildInfo("my_service", "1.0.0"))

	// First registration should succeed
	reg.NewCounter("requests_total", "Total requests")

	// Second registration with same name should panic
	assert.Panics(t, func() {
		reg.NewCounter("requests_total", "Total requests duplicate")
	}, "Duplicate metric registration should panic")
}

func TestRegistry_PrependServiceName(t *testing.T) {
	reg := NewRegistry(build.NewBuildInfo("test_service", "1.0.0"))

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple name", "requests", "test_service_requests"},
		{"with suffix", "requests_total", "test_service_requests_total"},
		{"with underscores", "http_request_duration", "test_service_http_request_duration"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reg.prependServiceName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig("my_service")

	assert.NotNil(t, cfg)
	assert.True(t, cfg.Enabled)
	assert.Equal(t, "my_service", cfg.ServiceName)
	assert.Equal(t, "/metrics", cfg.Path)
	assert.True(t, cfg.IncludeRuntime)
}

func TestRegistry_Unwrap(t *testing.T) {
	reg := NewRegistry(build.NewBuildInfo("test_service", "1.0.0"))

	promReg := reg.Unwrap()
	require.NotNil(t, promReg)

	// Should be able to use the unwrapped registry directly
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "direct_counter",
		Help: "Directly registered counter",
	})
	err := promReg.Register(counter)
	assert.NoError(t, err)
}
