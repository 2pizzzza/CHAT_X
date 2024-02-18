package chat_controllers

import (
	"github.com/gofiber/fiber/v2"
	auth_controllers "github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/auth-controllers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
)

func DeleteMessage(c *fiber.Ctx) error {
	// Проверка аутентификации
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	user, err := auth_controllers.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	var params struct {
		MessageIDs []uint `json:"message_ids"`
	}
	if err := c.BodyParser(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request payload"})
	}

	messages := []*models.Message{}
	if result := initializers.DB.Where("user_id = ? AND id IN (?)", user.ID, params.MessageIDs).Find(&messages); result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch messages"})
	}

	if len(messages) != len(params.MessageIDs) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "User can only delete their own messages"})
	}

	if result := initializers.DB.Where("id IN (?)", params.MessageIDs).Delete(&models.Message{}); result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to delete messages"})
	}

	return c.JSON(fiber.Map{"message": "Messages deleted successfully"})
}
