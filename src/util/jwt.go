package util

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"vidhya-service/src/model"
)

// Claims is the JWT payload issued at login. It intentionally carries just
// enough data for RBAC decisions without a DB round-trip per request.
type Claims struct {
	UserID   string     `json:"sub"`
	SchoolID string     `json:"schoolId,omitempty"`
	Role     model.Role `json:"role"`
	Email    string     `json:"email"`
	jwt.RegisteredClaims
}

// JWTIssuer issues and verifies HMAC-signed access tokens.
type JWTIssuer struct {
	secret   []byte
	issuer   string
	accessTL time.Duration
}

// NewJWTIssuer creates a JWTIssuer with the given signing secret, issuer name,
// and access-token lifetime.
func NewJWTIssuer(secret, issuer string, accessTokenDuration time.Duration) *JWTIssuer {
	return &JWTIssuer{secret: []byte(secret), issuer: issuer, accessTL: accessTokenDuration}
}

// GenerateAccessToken issues a signed JWT access token for the given user.
func (j *JWTIssuer) GenerateAccessToken(userID, schoolID, email string, role model.Role) (string, time.Time, error) {
	expiresAt := time.Now().Add(j.accessTL)
	claims := Claims{
		UserID:   userID,
		SchoolID: schoolID,
		Role:     role,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(j.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("util.GenerateAccessToken: %w", err)
	}
	return signed, expiresAt, nil
}

// ParseAccessToken validates and decodes an access token into Claims.
func (j *JWTIssuer) ParseAccessToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return j.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("util.ParseAccessToken: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("util.ParseAccessToken: invalid token")
	}
	return claims, nil
}
