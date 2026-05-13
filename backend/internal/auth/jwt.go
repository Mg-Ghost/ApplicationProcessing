package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	Login  string `json:"login"`
	jwt.RegisteredClaims
}

func GenerateToken(userID int64, role string, login ...string) (string, error) {
	l := ""
	if len(login) > 0 { l = login[0] }
	claims := &Claims{
		UserID: userID,
		Role:   role,
		Login:  l,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret()))
}

func ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret()), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func jwtSecret() string {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		// В продакшене ОБЯЗАТЕЛЬНО задать JWT_SECRET в .env
		// Дефолт только для локальной разработки
		if os.Getenv("GIN_MODE") == "release" {
			panic("JWT_SECRET must be set in production (GIN_MODE=release)")
		}
		return "local_dev_secret_change_in_prod!!"
	}
	if len(s) < 32 {
		panic("JWT_SECRET must be at least 32 characters long")
	}
	return s
}
