package controllers

import (
	"context"
	"fmt"
	"io"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8soperatorsv1alpha1 "k8soperators/pkg/apis/k8soperators/v1alpha1"
	"k8soperators/pkg/constants"
	"k8soperators/pkg/metrics"
	"k8soperators/pkg/server/utils"
	"net/http"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	managedNamespaceController = Controller{
		Name: "ManagedNamespaceController",
		Path: "/managednamespace",
		Mux:  http.NewServeMux(),
	}

	createdNameSpacesCounter = 	metrics.GetCreatedNamespacesCounter()
	failedCreateNamespaceCounterVec = metrics.GetFailedCreateNamespaceCounterVec()
)

type RequestNamespaceBody struct {
	User string
}

var mnLog = logf.Log.WithName(managedNamespaceController.Name)

func RequestNamespace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		failedCreateNamespaceCounterVec.WithLabelValues("non-post").Inc()
		http.Error(w, "Only accepting POST requests", http.StatusMethodNotAllowed)
		return
	}

	var body RequestNamespaceBody
	if err := utils.GetJson(w, r, &body); err != nil {
		mnLog.Info(err.Error())
		failedCreateNamespaceCounterVec.WithLabelValues("json").Inc()
		return
	}

	namespace := &corev1.Namespace{}
	namespace.SetName(body.User)
	namespace.SetLabels(map[string]string{constants.K8sOperatorsLabelKey: constants.ManagedNamespaceLabelValue})

	roleBinding := &rbacv1.RoleBinding{}
	roleBinding.SetName("ephemeral-namespace-binding")
	roleBinding.SetNamespace(namespace.Name)
	roleBinding.RoleRef = rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     "admin",
	}
	roleBinding.Subjects = []rbacv1.Subject{
		{
			Kind:      "Group",
			APIGroup:  "rbac.authorization.k8s.io",
			Name:      "system:serviceaccounts",
			Namespace: namespace.Name,
		},
		{
			Kind:      "User",
			APIGroup:  "rbac.authorization.k8s.io",
			Name:      body.User,
			Namespace: namespace.Name,
		},
	}

	managedNamespace := &k8soperatorsv1alpha1.ManagedNamespace{}
	managedNamespace.SetName(constants.ManagedNamespaceName)
	managedNamespace.SetNamespace(namespace.Name)
	managedNamespace.Spec.Owner = body.User

	if err := k8sClient.Create(context.TODO(), namespace); err != nil {
		mnLog.Info(fmt.Sprintf("Failed to create Namespace %s for user %s", namespace.Name, body.User))
		failedCreateNamespaceCounterVec.WithLabelValues("namespace").Inc()
		http.Error(w, fmt.Sprintf("Failed to create Namespace: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if err := k8sClient.Create(context.TODO(), roleBinding); err != nil {
		k8sClient.Delete(context.TODO(), namespace)
		mnLog.Info(fmt.Sprintf("Failed to create RoleBinding in namespace %s for user %s", namespace.Name, body.User))
		failedCreateNamespaceCounterVec.WithLabelValues("rolebinding").Inc()
		http.Error(w, fmt.Sprintf("Failed to create RoleBinding: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if err := k8sClient.Create(context.TODO(), managedNamespace); err != nil {
		k8sClient.Delete(context.TODO(), namespace)
		mnLog.Info(fmt.Sprintf("Failed to create ManagedNamespace in namespace %s for user %s", namespace.Name, body.User))
		failedCreateNamespaceCounterVec.WithLabelValues("managednamespace").Inc()
		http.Error(w, fmt.Sprintf("Failed to create ManagedNamespace: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	mnLog.Info(fmt.Sprintf("Created ManagedNamespace in namespace %s for user %s", namespace.Name, body.User))
	createdNameSpacesCounter.Inc()

	io.WriteString(w, fmt.Sprintf("Namespace created: %s", namespace.Name))
}

func init() {
	managedNamespaceController.Mux.HandleFunc("/create", RequestNamespace)

	registerController(&managedNamespaceController)
}
