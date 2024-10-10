package middleware

import (
	"net/http"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var ehLog = logf.Log.WithName("LoggingMiddleware")

type exampleHandler1 struct {
	nextHandler http.Handler
}

type exampleHandler2 struct {
	nextHandler http.Handler
}

func (h exampleHandler1) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ehLog.Info("PRE 1")
	h.nextHandler.ServeHTTP(w, req)
	ehLog.Info("POST 1")
}

func (h exampleHandler2) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ehLog.Info("PRE 2")
	h.nextHandler.ServeHTTP(w, req)
	ehLog.Info("POST 2")
}

func GetExampleMiddleware1() *Middleware {
	return &Middleware{
		Name: "example1",
		Function: func(h http.Handler) http.Handler{
			return exampleHandler1{nextHandler: h}
		},
	}
}

func GetExampleMiddleware2() *Middleware {
	return &Middleware{
		Name: "example2",
		Function: func(h http.Handler) http.Handler{
			return exampleHandler2{nextHandler: h}
		},
	}
}
