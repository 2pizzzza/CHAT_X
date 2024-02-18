package chat_controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/auth-controllers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
)

func UpdateMessage(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	user, err := auth_controllers.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	var params struct {
		MessageID uint   `json:"message_id"`
		Text      string `json:"text"`
	}
	if err := c.BodyParser(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request payload"})
	}

	message := &models.Message{}
	if result := initializers.DB.Where("user_id = ? AND id = ?", user.ID, params.MessageID).First(message); result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch message"})
	}

	if message.ID == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Message not found"})
	}

	message.Text = params.Text
	if result := initializers.DB.Save(&message); result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to update message"})
	}

	return c.JSON(fiber.Map{"message": "Message updated successfully"})
}
