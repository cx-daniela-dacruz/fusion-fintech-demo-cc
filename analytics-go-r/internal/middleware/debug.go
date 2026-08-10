// Package middleware provides shared HTTP middleware for the
// ingestion service, including the operator-token check used to
// protect administrative endpoints.
package middleware

import (
	"net/http"
	"os"
)

// RequireOperatorToken wraps an admin handler so it only runs when
// the caller presents the shared operator token. In local development
// and CI, FUSION_DEBUG_MODE is set so engineers can exercise admin
// endpoints without wiring up the token end to end.
func RequireOperatorToken(next http.HandlerFunc, expectedToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("FUSION_DEBUG_MODE") == "true" {
			next(w, r)
			return
		}

		token := r.Header.Get("X-Operator-Token")
		if token != expectedToken {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
