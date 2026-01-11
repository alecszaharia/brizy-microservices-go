package metrics

// Config holds metrics configuration.
type Config struct {
	Enabled        bool
	ServiceName    string
	Path           string
	IncludeRuntime bool
}

// DefaultConfig returns default metrics configuration.
func DefaultConfig(serviceName string) *Config {
	return &Config{
		Enabled:        true,
		ServiceName:    serviceName,
		Path:           "/metrics",
		IncludeRuntime: true,
	}
}
