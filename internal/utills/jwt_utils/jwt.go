package jwt_utils

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	initializers2 "github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"time"
)

func GenerateTokens(userID *uuid.UUID) (string, string, error) {
	config, err := initializers2.LoadConfig(".")
	if err != nil {
		return "", "", err
	}

	// Create access token
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

	// Create refresh token
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
