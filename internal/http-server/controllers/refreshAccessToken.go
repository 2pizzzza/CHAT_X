package controllers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	initializers2 "github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/utills/jwt_utils"
)

func RefreshAccessToken(c *fiber.Ctx) error {
	type RefreshTokenInput struct {
		RefreshToken string `json:"refresh_token"`
	}

	var input RefreshTokenInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	// Проверка наличия refresh token в теле запроса
	if input.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Refresh token is required"})
	}

	// Проверка валидности refresh token
	token, err := jwt.Parse(input.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		// Проверяем, используется ли правильный алгоритм для токена
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Возвращаем ключ для расшифровки токена
		config, err := initializers2.LoadConfig(".")
		if err != nil {
			return nil, err
		}
		return []byte(config.JwtSecret), nil
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Invalid refresh token"})
	}

	// Проверка валидности токена
	if !token.Valid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Invalid refresh token"})
	}

	// Проверка типа токена
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Invalid refresh token"})
	}

	// Получение идентификатора пользователя из токена
	userID, err := uuid.Parse(claims["sub"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "fail", "message": "Failed to parse user ID"})
	}

	// Генерация нового access token
	accessToken, _, err := jwt_utils.GenerateTokens(&userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "fail", "message": "Failed to generate access token"})
	}

	// Отправляем новый access token в ответе
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "access_token": accessToken})
}
