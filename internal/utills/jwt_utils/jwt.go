package jwt_utils

import (
	"errors"
	"fmt"
	"github.com/gofiber/contrib/websocket"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"log/slog"
	"strings"
	"time"
)

func GenerateTokens(userID *uuid.UUID) (string, string, error) {
	config, err := initializers.LoadConfig()
	if err != nil {
		return "", "", err
	}

	accessToken := jwt.New(jwt.SigningMethodHS256)
	accessClaims := accessToken.Claims.(jwt.MapClaims)
	accessClaims["sub"] = userID
	accessClaims["exp"] = time.Now().Add(config.JwtExpiresIn).Unix()
	accessClaims["iat"] = time.Now().Unix()
	accessClaims["nbf"] = time.Now().Unix()
	accessTokenString, err := accessToken.SignedString([]byte(config.JwtSecret))
	if err != nil {
		return "", "", fmt.Errorf("generating JWT Token failed: %v", err)
	}

	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshClaims["sub"] = userID
	refreshClaims["exp"] = time.Now().Add(config.RefreshTokenExpiresIn).Unix()
	refreshClaims["iat"] = time.Now().Unix()
	refreshTokenString, err := refreshToken.SignedString([]byte(config.JwtSecret))
	if err != nil {
		return "", "", fmt.Errorf("generating Refresh Token failed: %v", err)
	}

	return accessTokenString, refreshTokenString, nil
}

func ValidateToken(c *websocket.Conn) (jwt.MapClaims, error) {
	tokenString := c.Headers("Authorization")
	config, err := initializers.LoadConfig()
	jwtSecret := config.JwtSecret
	if err != nil {
		slog.Error("error load config is", err)
	}
	if tokenString == "" {
		return nil, errors.New("authorization header is missing")
	}

	tokenString = strings.Replace(tokenString, "Bearer ", "", 1)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {

		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
