// Package common общие понятия
package common

import "net/http"

func RequireHTTPMethodMiddleware(method string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != method {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
