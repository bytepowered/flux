package server

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
	// 请求访问次数统计
	EndpointAccess *prometheus.CounterVec
	// 请求错误次数统计
	EndpointError *prometheus.CounterVec
	// 请求耗时次数统计
	RouteDuration *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	// rer: https://prometheus.io/docs/concepts/data_model/
	// must match the regex [a-zA-Z_:][a-zA-Z0-9_:]*.
	const namespace, subsystem = "fluxgo", "runtime"
	return &Metrics{
		EndpointAccess: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "access_count",
			Help:      "Number of endpoint access",
		}, []string{"Listener", "Method", "Pattern", "Version"}),
		EndpointError: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "error_count",
			Help:      "Number of endpoint access errors",
		}, []string{"Listener", "Method", "Pattern", "Version", "ErrorCode"}),
		RouteDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "route_duration",
			Help:      "Spend time by processing a endpoint",
			Buckets:   defaultMetricBuckets,
		}, []string{"ComponentType", "TypeId"}),
	}
}
