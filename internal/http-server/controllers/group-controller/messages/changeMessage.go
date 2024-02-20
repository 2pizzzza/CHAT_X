package messages

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/auth-controllers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
)

func UpdateGroupMessage(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	user, err := auth_controllers.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	var updateMessageRequest struct {
		GroupID   uint   `json:"group_id"`
		MessageID uint   `json:"message_id"`
		Text      string `json:"text"`
	}
	if err := c.BodyParser(&updateMessageRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request"})
	}

	var message models.GroupMessage
	if err := initializers.DB.Where("id = ? AND group_id = ?", updateMessageRequest.MessageID, updateMessageRequest.GroupID).First(&message).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Message not found"})
	}

	if message.UserID.String() != user.ID.String() {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "You are not allowed to update this message"})
	}

	if err := initializers.DB.Model(&message).Update("text", updateMessageRequest.Text).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to update message"})
	}

	return c.JSON(fiber.Map{"message": "Message updated successfully", "data": models.FilterGroupMessageRecord(&message)})
}
