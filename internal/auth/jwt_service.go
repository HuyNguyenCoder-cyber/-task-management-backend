package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secret         []byte
	expiresInHours int
}

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

var ErrInvalidToken = errors.New("invalid token")

func NewJWTService(secret string, expiresInHours int) *JWTService {
	return &JWTService{
		secret:         []byte(secret),
		expiresInHours: expiresInHours,
	}
}

func (s *JWTService) GenerateAccessToken(userID int64) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(s.expiresInHours) * time.Hour)
	expiresIn := int64(time.Until(expiresAt).Seconds())

	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.secret)
	if err != nil {
		return "", 0, err
	}

	return signedToken, expiresIn, nil
}

func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}

		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if token == nil || !token.Valid || claims.UserID <= 0 {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
