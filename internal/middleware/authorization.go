package middleware

import "net/http"


func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := UserRoleFromContext(r.Context())

		if !ok || role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return 
		}

		next.ServeHTTP(w, r)
	})
}