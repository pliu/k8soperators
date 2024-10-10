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
	controllers.RegisterConfig(mgr.GetConfig())
	routes := registerControllers()
	log.Info("Global middleware")
	rootMux.Handle("/", chainMiddleware(routes, []*middleware.Middleware{
		middleware.GetLoggingMiddleware(),
	}))

	go http.ListenAndServe(address, rootMux)

	log.Info(fmt.Sprintf("K8sOperators server listening at %s", address))
}

func registerControllers() *http.ServeMux {
	routes := http.NewServeMux()
	for _, controller := range controllers.Controllers {
		log.Info(fmt.Sprintf("%s registered at %s", controller.Name, controller.Path))
		routes.Handle(fmt.Sprintf("%s/", controller.Path), http.StripPrefix(controller.Path, chainMiddleware(controller.Mux, controller.Middlewares)))
	}
	return routes
}

func chainMiddleware(handler http.Handler, middlewares []*middleware.Middleware) http.Handler {
	handlerBuilder := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		log.Info(fmt.Sprintf("Using %s middleware", middlewares[i].Name))
		handlerBuilder = middlewares[i].Function(handlerBuilder)
	}
	return handlerBuilder
}
