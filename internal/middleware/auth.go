package middleware

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rahul810050/go-store/internal/auth"
)

type contextType string

const (
	userIDKey contextType = "userID"
	userRoleKey contextType = "userRole"
)

func UserIDFromContext(ctx context.Context) (int64, bool) {
    id, ok := ctx.Value(userIDKey).(int64)
    return id, ok
}

func UserRoleFromContext(ctx context.Context) (string, bool) {
    role, ok := ctx.Value(userRoleKey).(string)
    return role, ok
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		header := r.Header.Get("Authorization")
		parts := strings.SplitN(header, " ", 2)

		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

		secret := os.Getenv("JWT_SECRET")

		claims := &auth.Claims{}

		token, err := jwt.ParseWithClaims(
            parts[1],
            claims,
            func(token *jwt.Token) (any, error) {
                return []byte(secret), nil
            },
            jwt.WithValidMethods([]string{"HS256"}),
            jwt.WithIssuer("ecom-api"),
            jwt.WithExpirationRequired(),
        )

		if err != nil || !token.Valid {
            http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
            return
        }
		// extracting userid
		userID, err := strconv.ParseInt(claims.Subject, 10, 64)
        if err != nil || userID <= 0 {
            http.Error(w, "Invalid token claims", http.StatusUnauthorized)
            return
        }

		if claims.Role != "user" && claims.Role != "admin" {
            http.Error(w, "Invalid token claims", http.StatusUnauthorized)
            return
        }

		ctx := context.WithValue(r.Context(), userIDKey, userID)
        ctx = context.WithValue(ctx, userRoleKey, claims.Role)

        next.ServeHTTP(w, r.WithContext(ctx))
	})
}