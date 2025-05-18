package middleware

import (
	"context"
	"net/http"

	"github.com/dgrijalva/jwt-go"

	"github.com/vikasavnish/trademicro/internal/models"
)

// AuthMiddleware checks for valid JWT token and adds username to context
func AuthMiddleware(jwtSecretKey []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorizationHeader := r.Header.Get("Authorization")
			if authorizationHeader == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Extract the token
			tokenString := authorizationHeader[7:] // Remove "Bearer " prefix

			// Parse and validate the token
			claims := &models.Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return jwtSecretKey, nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Add the username to the request context
			ctx := context.WithValue(r.Context(), "username", claims.Username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
