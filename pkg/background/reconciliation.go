package background

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type GlobalReconciler func(c client.Client) bool

var (
	globalReconcilers []GlobalReconciler
	log               = logf.Log.WithName("initialization")
)

func AddGlobalReconciler(globalReconciler GlobalReconciler) {
	globalReconcilers = append(globalReconcilers, globalReconciler)
}

func InitialReconciliation(c client.Client) {
	log.Info("Starting initial global reconciliations")
	for _, globalReconciler := range globalReconcilers {
		globalReconciler(c)
	}
}
