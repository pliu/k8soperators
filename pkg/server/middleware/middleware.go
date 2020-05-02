package middleware

import "net/http"

type middleware func(handler http.Handler) http.Handler

func ChainMiddleware(handler http.Handler, middlewares []middleware) http.Handler {
	handlerBuilder := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		handlerBuilder = middlewares[i](handlerBuilder)
	}
	return handlerBuilder
}

func GetRootMiddleware(handler http.Handler) http.Handler {
	middlewares := []middleware{
		exampleMiddleware1,
		exampleMiddleware2,
		loggingMiddleware,
	}
	return ChainMiddleware(handler, middlewares)
}
