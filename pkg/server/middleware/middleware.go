package middleware

import "net/http"

func ChainMiddleware(handlers []*http.Handler) http.Handler {
	return nil
}

func GetUniversalMiddleware() http.Handler {
	return nil
}
