package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// Registry provides methods for creating custom metrics with proper namespacing.
type Registry struct {
	registry    *prometheus.Registry
	serviceName string
}

// NewRegistry creates a new metrics registry for the given service.
// It automatically registers Go runtime collectors and build info.
func NewRegistry(serviceName string) *Registry {
	reg := prometheus.NewRegistry()

	// Register Go runtime collectors
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	// Register build info
	registerBuildInfo(reg, serviceName)

	return &Registry{
		registry:    reg,
		serviceName: serviceName,
	}
}

// Unwrap returns the underlying Prometheus registry for direct access.
func (r *Registry) Unwrap() *prometheus.Registry {
	return r.registry
}

// prependServiceName prepends the service name to metric names.
func (r *Registry) prependServiceName(name string) string {
	return r.serviceName + "_" + name
}

// NewCounterVec creates a counter with labels.
// The service name is automatically prepended to the metric name.
func (r *Registry) NewCounterVec(name, help string, labelNames []string) *prometheus.CounterVec {
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: r.prependServiceName(name),
			Help: help,
		},
		labelNames,
	)
	r.registry.MustRegister(counter)
	return counter
}

// NewGaugeVec creates a gauge with labels.
// The service name is automatically prepended to the metric name.
func (r *Registry) NewGaugeVec(name, help string, labelNames []string) *prometheus.GaugeVec {
	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: r.prependServiceName(name),
			Help: help,
		},
		labelNames,
	)
	r.registry.MustRegister(gauge)
	return gauge
}

// NewHistogramVec creates a histogram with custom buckets and labels.
// The service name is automatically prepended to the metric name.
func (r *Registry) NewHistogramVec(name, help string, buckets []float64, labelNames []string) *prometheus.HistogramVec {
	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    r.prependServiceName(name),
			Help:    help,
			Buckets: buckets,
		},
		labelNames,
	)
	r.registry.MustRegister(histogram)
	return histogram
}

// NewCounter creates a counter without labels (convenience method).
// The service name is automatically prepended to the metric name.
func (r *Registry) NewCounter(name, help string) prometheus.Counter {
	counter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: r.prependServiceName(name),
			Help: help,
		},
	)
	r.registry.MustRegister(counter)
	return counter
}

// NewGauge creates a gauge without labels (convenience method).
// The service name is automatically prepended to the metric name.
func (r *Registry) NewGauge(name, help string) prometheus.Gauge {
	gauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: r.prependServiceName(name),
			Help: help,
		},
	)
	r.registry.MustRegister(gauge)
	return gauge
}
