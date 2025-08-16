package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService struct {
	secretKey []byte
	issuer    string
}

type Claims struct {
	UserID         uuid.UUID `json:"user_id"`
	Email          string    `json:"email"`
	Rol            string    `json:"rol"`
	OrganizacionID *uuid.UUID `json:"organizacion_id,omitempty"`
	jwt.RegisteredClaims
}

func NewJWTService(secretKey, issuer string) *JWTService {
	return &JWTService{
		secretKey: []byte(secretKey),
		issuer:    issuer,
	}
}

func (s *JWTService) GenerateToken(userID uuid.UUID, email, rol string, organizacionID *uuid.UUID) (string, error) {
	claims := Claims{
		UserID:         userID,
		Email:          email,
		Rol:            rol,
		OrganizacionID: organizacionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 horas
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    s.issuer,
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("token inválido")
}

func (s *JWTService) RefreshToken(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Generar nuevo token
	return s.GenerateToken(claims.UserID, claims.Email, claims.Rol, claims.OrganizacionID)
}