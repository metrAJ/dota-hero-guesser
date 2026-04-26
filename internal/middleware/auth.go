package middleware

import (
	"context"
	"doproj/internal/auth"

	"net/http"
	"strings"
)

type ContextKey string

const UserIDContextKey ContextKey = "userID"

type AuthMiddleware struct {
	tokenManager *auth.TokenManager
}

func NewAuthMiddleware(tm *auth.TokenManager) *AuthMiddleware {
	return &AuthMiddleware{
		tokenManager: tm,
	}
}

func (m *AuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract the token from the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}
		// Expecting header format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}
		// Validate the token and extract the user ID
		userID, err := m.tokenManager.ValidateToken(parts[1])
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		// Store the user ID in the request context for downstream handlers
		ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
		// Call the next handler with the updated context
		next(w, r.WithContext(ctx))
	}
}
