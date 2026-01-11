package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_DefaultConfig(t *testing.T) {
	cfg := DefaultConfig("test_service")

	assert.NotNil(t, cfg)
	assert.True(t, cfg.Enabled, "Metrics should be enabled by default")
	assert.Equal(t, "test_service", cfg.ServiceName)
	assert.Equal(t, "/metrics", cfg.Path, "Default path should be /metrics")
	assert.True(t, cfg.IncludeRuntime, "Runtime metrics should be included by default")
}

func TestConfig_CustomValues(t *testing.T) {
	cfg := &Config{
		Enabled:        false,
		ServiceName:    "custom_service",
		Path:           "/custom_metrics",
		IncludeRuntime: false,
	}

	assert.False(t, cfg.Enabled)
	assert.Equal(t, "custom_service", cfg.ServiceName)
	assert.Equal(t, "/custom_metrics", cfg.Path)
	assert.False(t, cfg.IncludeRuntime)
}

func TestConfig_MultipleServices(t *testing.T) {
	services := []string{"users", "products", "orders"}

	for _, svc := range services {
		cfg := DefaultConfig(svc)
		assert.Equal(t, svc, cfg.ServiceName, "Service name should match input")
		assert.True(t, cfg.Enabled)
		assert.Equal(t, "/metrics", cfg.Path)
	}
}
