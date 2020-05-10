package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var collectors []prometheus.Collector

func addCollector(collector prometheus.Collector) {
	collectors = append(collectors, collector)
}

func RegisterCollectors() {
	for _, collector := range collectors {
		metrics.Registry.MustRegister(collector)
	}
}
