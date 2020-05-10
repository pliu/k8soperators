package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	createdNamespacesCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "managednamespace_created_namespaces",
		Help: "Number of namespaces creates",
	})

	failedCreateNamespaceCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "managednamespace_failed__create_namespace",
		Help: "Number of namespace creation failures",
	}, []string{"reason"})

	deletedNamespacesCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "managednamespace_deleted_namespaces",
		Help: "Number of namespaces deleted",
	})
)

func GetCreatedNamespacesCounter() prometheus.Counter {
	return createdNamespacesCounter
}

func GetFailedCreateNamespaceCounterVec() *prometheus.CounterVec {
	return failedCreateNamespaceCounterVec
}

func GetDeletedNamespacesCounter() prometheus.Counter {
	return deletedNamespacesCounter
}

func init() {
	addCollector(createdNamespacesCounter)
	addCollector(failedCreateNamespaceCounterVec)
	addCollector(deletedNamespacesCounter)
}
