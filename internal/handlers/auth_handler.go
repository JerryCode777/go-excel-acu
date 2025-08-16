package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"goexcel/internal/auth"
	"goexcel/internal/database/repositories"
	"goexcel/internal/models"

	"github.com/google/uuid"
)

type AuthHandler struct {
	usuarioRepo *repositories.UsuarioRepository
	jwtService  *auth.JWTService
}

func NewAuthHandler(usuarioRepo *repositories.UsuarioRepository, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		usuarioRepo: usuarioRepo,
		jwtService:  jwtService,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.UsuarioCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Datos inválidos")
		return
	}

	// Verificar si el email ya existe
	exists, err := h.usuarioRepo.EmailExists(req.Email)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error verificando email")
		return
	}
	if exists {
		h.respondWithError(w, http.StatusConflict, "El email ya está registrado")
		return
	}

	// Hash de la contraseña
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error procesando contraseña")
		return
	}

	// Crear usuario
	usuario := &models.Usuario{
		ID:             uuid.New(),
		Email:          req.Email,
		PasswordHash:   passwordHash,
		Nombre:         req.Nombre,
		Apellido:       req.Apellido,
		Rol:            "user", // Rol por defecto
		OrganizacionID: req.OrganizacionID,
		Activo:         true,
	}

	if err := h.usuarioRepo.Create(usuario); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error creando usuario")
		return
	}

	// Generar token
	token, err := h.jwtService.GenerateToken(usuario.ID, usuario.Email, usuario.Rol, usuario.OrganizacionID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error generando token")
		return
	}

	// Actualizar último acceso
	h.usuarioRepo.UpdateLastAccess(usuario.ID)

	// Respuesta sin password hash
	usuario.PasswordHash = ""
	response := models.UsuarioLoginResponse{
		Token:   token,
		Usuario: *usuario,
	}

	h.respondWithJSON(w, http.StatusCreated, response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.UsuarioLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Datos inválidos")
		return
	}

	// Buscar usuario por email
	usuario, err := h.usuarioRepo.GetByEmail(req.Email)
	if err != nil {
		h.respondWithError(w, http.StatusUnauthorized, "Credenciales inválidas")
		return
	}

	// Verificar contraseña
	if !auth.CheckPassword(req.Password, usuario.PasswordHash) {
		h.respondWithError(w, http.StatusUnauthorized, "Credenciales inválidas")
		return
	}

	// Generar token
	token, err := h.jwtService.GenerateToken(usuario.ID, usuario.Email, usuario.Rol, usuario.OrganizacionID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error generando token")
		return
	}

	// Actualizar último acceso
	h.usuarioRepo.UpdateLastAccess(usuario.ID)

	// Respuesta sin password hash
	usuario.PasswordHash = ""
	response := models.UsuarioLoginResponse{
		Token:   token,
		Usuario: *usuario,
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		h.respondWithError(w, http.StatusUnauthorized, "Usuario no autenticado")
		return
	}

	// Generar nuevo token
	token, err := h.jwtService.GenerateToken(user.ID, user.Email, user.Rol, user.OrganizacionID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error generando token")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		h.respondWithError(w, http.StatusUnauthorized, "Usuario no autenticado")
		return
	}

	// Obtener información completa del usuario
	fullUser, err := h.usuarioRepo.GetByID(user.ID)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "Usuario no encontrado")
		return
	}

	// Limpiar datos sensibles
	fullUser.PasswordHash = ""

	h.respondWithJSON(w, http.StatusOK, fullUser)
}

func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		h.respondWithError(w, http.StatusUnauthorized, "Usuario no autenticado")
		return
	}

	var req models.UsuarioUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Datos inválidos")
		return
	}

	// Obtener usuario actual
	usuario, err := h.usuarioRepo.GetByID(user.ID)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "Usuario no encontrado")
		return
	}

	// Actualizar campos
	if req.Nombre != nil {
		usuario.Nombre = *req.Nombre
	}
	if req.Apellido != nil {
		usuario.Apellido = req.Apellido
	}
	if req.AvatarURL != nil {
		usuario.AvatarURL = req.AvatarURL
	}
	if req.OrganizacionID != nil {
		usuario.OrganizacionID = req.OrganizacionID
	}

	if err := h.usuarioRepo.Update(usuario); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error actualizando perfil")
		return
	}

	// Limpiar datos sensibles
	usuario.PasswordHash = ""

	h.respondWithJSON(w, http.StatusOK, usuario)
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		h.respondWithError(w, http.StatusUnauthorized, "Usuario no autenticado")
		return
	}

	var req models.UsuarioChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Datos inválidos")
		return
	}

	// Obtener usuario actual con hash de contraseña
	usuario, err := h.usuarioRepo.GetByID(user.ID)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "Usuario no encontrado")
		return
	}

	// Verificar contraseña actual
	if !auth.CheckPassword(req.CurrentPassword, usuario.PasswordHash) {
		h.respondWithError(w, http.StatusUnauthorized, "Contraseña actual incorrecta")
		return
	}

	// Hash de la nueva contraseña
	newPasswordHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error procesando nueva contraseña")
		return
	}

	// Actualizar contraseña
	if err := h.usuarioRepo.UpdatePassword(user.ID, newPasswordHash); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error actualizando contraseña")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Contraseña actualizada exitosamente"})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// En una implementación completa, aquí se invalidaría el token en una blacklist
	// Por ahora solo respondemos success ya que JWT es stateless
	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Logout exitoso"})
}

// Helper methods
func (h *AuthHandler) respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":     message,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (h *AuthHandler) respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Printf("Error encoding JSON response: %v\n", err)
	}
}