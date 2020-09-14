package handlers

import (
	"context"
	"linkener/internal/db"
	"net/http"
)

// AuthMiddleware - ensure valid access token is passed for API routes that require authentication
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized access", http.StatusUnauthorized)
			return
		}

		stmt, err := db.DBCon.Prepare("select username from access_tokens where access_token=? and expiry>CURRENT_TIMESTAMP limit 1")
		if err != nil {
			println(err.Error())
			http.Error(w, "Failed to authorize request", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		rows, err := stmt.Query(token)
		if err != nil {
			println(err.Error())
			http.Error(w, "Failed to authorize request", http.StatusInternalServerError)
			return
		}

		if rows.Next() {
			var username string
			rows.Scan(&username)
			rows.Close()
			ctxt := r.WithContext(context.WithValue(r.Context(), UsernameContextKey, username))
			next.ServeHTTP(w, ctxt)
		} else {
			rows.Close()
			http.Error(w, "Unauthorized access", http.StatusUnauthorized)
			return
		}
	})
}
