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
	EndpointAccess *prometheus.CounterVec
	EndpointError  *prometheus.CounterVec
	RouteDuration  *prometheus.HistogramVec
}

func NewMetrics(listenerId string) *Metrics {
	// rer: https://prometheus.io/docs/concepts/data_model/
	// must match the regex [a-zA-Z_:][a-zA-Z0-9_:]*.
	namespace := "onlistener:" + listenerId
	return &Metrics{
		EndpointAccess: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "endpoint",
			Name:      "access_count",
			Help:      "Number of endpoint access",
		}, []string{"ProtoName", "Interface", "Method"}),
		EndpointError: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "endpoint",
			Name:      "error_count",
			Help:      "Number of endpoint access errors",
		}, []string{"ProtoName", "Interface", "Method", "ErrorCode"}),
		RouteDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "endpoint",
			Name:      "route_duration",
			Help:      "Spend time by processing a endpoint",
			Buckets:   defaultMetricBuckets,
		}, []string{"ComponentType", "TypeId"}),
	}
}
