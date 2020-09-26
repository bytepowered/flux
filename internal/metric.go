package internal

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	defaultMetricNamespace = "flux"
	defaultMetricSubsystem = "http"
	defaultMetricBuckets   = []float64{
		0.0005,
		0.001, // 1ms
		0.002,
		0.005,
		0.01, // 10ms
		0.02,
		0.05,
		0.1, // 100 ms
		0.2,
		0.5,
		1.0, // 1s
		2.0,
		5.0,
		10.0, // 10s
		15.0,
		20.0,
		30.0,
	}
)

type Metrics struct {
	EndpointAccess *prometheus.CounterVec
	EndpointError  *prometheus.CounterVec
	RouteDuration  *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	return &Metrics{
		EndpointAccess: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: defaultMetricNamespace,
			Subsystem: defaultMetricSubsystem,
			Name:      "endpoint_access_total",
			Help:      "Number of endpoint access",
		}, []string{"ProtoName", "UpstreamUri", "UpstreamMethod"}),
		EndpointError: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: defaultMetricNamespace,
			Subsystem: defaultMetricSubsystem,
			Name:      "endpoint_error_total",
			Help:      "Number of endpoint access errors",
		}, []string{"ProtoName", "UpstreamUri", "UpstreamMethod", "ErrorCode"}),
		RouteDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: defaultMetricNamespace,
			Subsystem: defaultMetricSubsystem,
			Name:      "endpoint_route_duration",
			Help:      "Spend time by processing a endpoint",
			Buckets:   defaultMetricBuckets,
		}, []string{"ComponentType", "TypeId"}),
	}
}
