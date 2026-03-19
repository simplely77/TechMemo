package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType 标识token类型
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims JWT 声明结构
type Claims struct {
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// GenerateAccessToken 生成 accessToken
func GenerateAccessToken(userID int64, username string, secret string, expireHours int) (string, error) {
	return generateToken(userID, username, secret, time.Hour*time.Duration(expireHours), AccessToken)
}

// GenerateRefreshToken 生成 refreshToken
func GenerateRefreshToken(userID int64, username string, secret string, expireDays int) (string, error) {
	return generateToken(userID, username, secret, time.Hour*24*time.Duration(expireDays), RefreshToken)
}

// generateToken 生成 JWT token
func generateToken(userID int64, username string, secret string, duration time.Duration, tokenType TokenType) (string, error) {
	claims := Claims{
		UserID:    userID,
		Username:  username,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 解析 JWT token
func ParseToken(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
