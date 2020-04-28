package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	v1 "k8s.io/api/core/v1"
	k8soperatorsv1alpha1 "k8soperators/pkg/apis/k8soperators/v1alpha1"
	"k8soperators/pkg/constants"
	"net/http"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var managedNamespaceController = Controller{
	Name: "ManagedNamespaceController",
	Path: "/managednamespace",
	Mux:  http.NewServeMux(),
}

type RequestNamespaceBody struct {
	User string
}

var mnLog = logf.Log.WithName(managedNamespaceController.Name)

func RequestNamespace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		mnLog.Info("Received non-POST: %s", r.Method)
		http.Error(w, "Only accepting POST requests", http.StatusMethodNotAllowed)
		return
	}

	var body RequestNamespaceBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		mnLog.Info(fmt.Sprintf("Invalid body: %s", err.Error()))
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	mnLog.Info(fmt.Sprintf("%s requesting namespace", body.User))

	namespace := &v1.Namespace{}
	namespace.SetName(body.User)
	namespace.SetLabels(map[string]string{constants.K8sOperatorsLabelKey: constants.ManagedNamespaceLabelValue})

	managedNamespace := &k8soperatorsv1alpha1.ManagedNamespace{}
	managedNamespace.SetName("manager")
	managedNamespace.SetNamespace(namespace.Name)
	managedNamespace.Spec.Owner = body.User

	err = k8sClient.Create(context.TODO(), namespace)
	if err != nil {
		mnLog.Info(fmt.Sprintf("Failed to create Namespace %s for user %s", namespace.Name, body.User))
		http.Error(w, fmt.Sprintf("Failed to create Namespace: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	mnLog.Info(fmt.Sprintf("Created Namespace %s for user %s", namespace.Name, body.User))

	err = k8sClient.Create(context.TODO(), managedNamespace)
	if err != nil {
		k8sClient.Delete(context.TODO(), namespace)
		mnLog.Info(fmt.Sprintf("Failed to create ManagedNamespace in namespace %s for user %s", namespace.Name, body.User))
		http.Error(w, fmt.Sprintf("Failed to create ManagedNamespace: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	mnLog.Info(fmt.Sprintf("Created ManagedNamespace in namespace %s for user %s", namespace.Name, body.User))

	io.WriteString(w, fmt.Sprintf("Namespace created: %s", namespace.Name))
}

func init() {
	managedNamespaceController.Mux.HandleFunc("/create", RequestNamespace)

	registerController(&managedNamespaceController)
}
