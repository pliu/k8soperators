package server

import (
	"context"
	"fmt"
	"io"
	v1 "k8s.io/api/core/v1"
	"net/http"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	log                     = logf.Log.WithName("k8soperators server")
	k8sClient client.Client = nil
)

func StartServer(mgr manager.Manager, address string) {
	k8sClient = mgr.GetClient()

	mux := http.NewServeMux()
	registerHandlers(mux)
	
	err := http.ListenAndServe(address, mux)
	if err != nil {
		log.Error(err, "K8sOperators failed to start")
		os.Exit(1)
	}
	log.Info(fmt.Sprintf("K8sOperators server listening at %s", address))
}

func registerHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/", ExampleHandler)
}

func ExampleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	namespaces := &v1.NamespaceList{}
	err := k8sClient.List(context.TODO(), namespaces)
	if err != nil {

	}
	total := ""
	for _, namespace := range namespaces.Items {
		total += namespace.Name + " "
	}
	io.WriteString(w, total)
}
