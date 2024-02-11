package controllers

import (
	"github.com/gofiber/fiber/v2"
	initializers2 "github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"golang.org/x/crypto/bcrypt"
)

var ChangePasswordRequest struct {
	CurrentPassword string `json:"old_password"`
	NewPassword     string `json:"new_password"`
}

func ChangePassword(c *fiber.Ctx) error {
	type ChangePasswordInput struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}

	// Парсинг входных данных
	var input ChangePasswordInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	// Получение пользователя из токена
	user, err := getUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "Unauthorized"})
	}

	// Проверка текущего пароля
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.CurrentPassword))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "Incorrect current password"})
	}

	// Хэширование нового пароля
	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "fail", "message": "Failed to hash new password"})
	}

	// Обновление пароля в базе данных
	user.Password = string(hashedNewPassword)
	if err := initializers2.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "fail", "message": "Failed to update password"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Password changed successfully"})
}
