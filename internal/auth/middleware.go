package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"goexcel/internal/models"
)

type contextKey string

const UserContextKey contextKey = "user"

type AuthMiddleware struct {
	jwtService *JWTService
}

func NewAuthMiddleware(jwtService *JWTService) *AuthMiddleware {
	return &AuthMiddleware{jwtService: jwtService}
}

func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := m.extractToken(r)
		if token == "" {
			m.respondWithError(w, http.StatusUnauthorized, "Token de acceso requerido")
			return
		}

		claims, err := m.jwtService.ValidateToken(token)
		if err != nil {
			m.respondWithError(w, http.StatusUnauthorized, "Token inválido")
			return
		}

		// Crear usuario básico desde claims
		user := &models.Usuario{
			ID:             claims.UserID,
			Email:          claims.Email,
			Rol:            claims.Rol,
			OrganizacionID: claims.OrganizacionID,
		}

		// Agregar usuario al contexto
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (m *AuthMiddleware) RequireRole(roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return m.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r.Context())
			if user == nil {
				m.respondWithError(w, http.StatusUnauthorized, "Usuario no autenticado")
				return
			}

			// Verificar rol
			hasRole := false
			for _, role := range roles {
				if user.Rol == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				m.respondWithError(w, http.StatusForbidden, "No tienes permisos para acceder a este recurso")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *AuthMiddleware) OptionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := m.extractToken(r)
		if token != "" {
			claims, err := m.jwtService.ValidateToken(token)
			if err == nil {
				user := &models.Usuario{
					ID:             claims.UserID,
					Email:          claims.Email,
					Rol:            claims.Rol,
					OrganizacionID: claims.OrganizacionID,
				}
				ctx := context.WithValue(r.Context(), UserContextKey, user)
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	}
}

func (m *AuthMiddleware) extractToken(r *http.Request) string {
	// Intentar extraer desde header Authorization
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Intentar extraer desde cookie
	cookie, err := r.Cookie("auth_token")
	if err == nil {
		return cookie.Value
	}

	return ""
}

func (m *AuthMiddleware) respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func GetUserFromContext(ctx context.Context) *models.Usuario {
	if user, ok := ctx.Value(UserContextKey).(*models.Usuario); ok {
		return user
	}
	return nil
}