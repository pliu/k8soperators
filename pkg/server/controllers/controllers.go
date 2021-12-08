package controllers

import (
	"fmt"
	"net/http"
	"os"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type Controller struct {
	Name string
	Path string
	Mux  *http.ServeMux
}

var (
	Controllers               = make(map[string]*Controller)
	k8sClient   client.Client = nil
	k8sConfig   *rest.Config  = nil
)

func registerController(controller *Controller) {
	log := logf.Log.WithName("controllers")
	if _, exists := Controllers[controller.Path]; exists {
		log.Error(fmt.Errorf(""), fmt.Sprintf("Two controllers with the same path %s", controller.Path))
		os.Exit(1)
	}
	Controllers[controller.Path] = controller
}

func RegisterClient(client client.Client) {
	k8sClient = client
}

func RegisterConfig(config *rest.Config) {
	k8sConfig = config
}
