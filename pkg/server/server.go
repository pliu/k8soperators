package server

import (
	"fmt"
	"k8soperators/pkg/server/controllers"
	"k8soperators/pkg/server/middleware"
	"net/http"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	log     = logf.Log.WithName("server")
	rootMux = http.NewServeMux()
)

func StartServer(mgr manager.Manager, address string) {
	controllers.RegisterClient(mgr.GetClient())
	routes := registerControllers()
	rootMux.Handle("/", middleware.GetRootMiddleware(routes))

	go http.ListenAndServe(address, rootMux)

	log.Info(fmt.Sprintf("K8sOperators server listening at %s", address))
}

func registerControllers() *http.ServeMux {
	routes := http.NewServeMux()
	for _, controller := range controllers.Controllers {
		log.Info(fmt.Sprintf("%s registered at %s", controller.Name, controller.Path))
		routes.Handle(fmt.Sprintf("%s/", controller.Path), http.StripPrefix(controller.Path, controller.Mux))
	}
	return routes
}
