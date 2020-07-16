package managednamespace

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	k8soperatorsv1alpha1 "k8soperators/pkg/apis/k8soperators/v1alpha1"
	"k8soperators/pkg/constants"
	"k8soperators/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type void struct{}
var (
	voidValue void
	log                      = logf.Log.WithName("controller_managednamespace")
	deletedNamespacesCounter = metrics.GetDeletedNamespacesCounter()
)

// Add creates a new ManagedNamespace Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileManagedNamespace{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("managednamespace-controller", mgr, controller.Options{
		Reconciler: r,
		MaxConcurrentReconciles: 1,
	})
	if err != nil {
		return err
	}

	// Watch for ManagedNamespace deletion events
	src := &source.Kind{Type: &k8soperatorsv1alpha1.ManagedNamespace{}}
	h := &handler.EnqueueRequestForObject{}
	pred := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			log.Info("Create - no action")
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			log.Info("Update - no action")
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			log.Info("Generic - no action")
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			log.Info("Delete")
			return true
		},
	}
	if err := c.Watch(src, h, pred); err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileManagedNamespace implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileManagedNamespace{}

// ReconcileManagedNamespace reconciles a ManagedNamespace object
type ReconcileManagedNamespace struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ManagedNamespace object and makes changes based on the state read
// and what is in the ManagedNamespace.Spec
func (r *ReconcileManagedNamespace) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ManagedNamespace")

	managedNamespaces := &k8soperatorsv1alpha1.ManagedNamespaceList{}
	if err := r.client.List(context.TODO(), managedNamespaces); err != nil {
		reqLogger.Error(err, "Failed to get ManagedNamepsaces - requeuing")
		return reconcile.Result{}, err
	}
	managedNamespaceNames := make(map[string]void)
	for _, managedNamespace := range managedNamespaces.Items {
		managedNamespaceNames[managedNamespace.Namespace] = voidValue
	}

	namespaces := &corev1.NamespaceList{}
	labelSelector := labels.SelectorFromSet(map[string]string{
		constants.K8sOperatorsLabelKey: constants.ManagedNamespaceLabelValue,
	})
	listOps := &client.ListOptions{LabelSelector: labelSelector}
	if err := r.client.List(context.TODO(), namespaces, listOps); err != nil {
		reqLogger.Error(err, "Failed to get Namepsaces - requeuing")
		return reconcile.Result{}, err
	}

	hasFailures := false
	for _, namespace := range namespaces.Items {
		if _, exists := managedNamespaceNames[namespace.Name]; exists {
			reqLogger.Info(fmt.Sprintf("Namespace %s is still being managed", namespace.Name))
			continue
		}
		if err := r.client.Delete(context.TODO(), &namespace); err != nil {
			reqLogger.Error(err, fmt.Sprintf("Failed to delete %s - requeuing", namespace.Name))
			hasFailures = true
		}
		reqLogger.Info(fmt.Sprintf("Deleted namespace %s", namespace.Name))
		deletedNamespacesCounter.Inc()
	}

	if hasFailures {
		return reconcile.Result{Requeue: true}, nil
	}

	time.Sleep(time.Second * 15)
	return reconcile.Result{}, nil
}
