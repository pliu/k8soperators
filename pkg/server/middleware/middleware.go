package middleware

import "net/http"

type Middleware struct {
	Name     string
	Function func(handler http.Handler) http.Handler
}
