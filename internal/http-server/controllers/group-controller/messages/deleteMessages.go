package messages

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/auth-controllers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
)

func DeleteGroupMessages(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	user, err := auth_controllers.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	var deleteMessagesRequest struct {
		GroupID    uint   `json:"group_id"`
		MessageIDs []uint `json:"message_ids"`
	}
	if err := c.BodyParser(&deleteMessagesRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request"})
	}

	var group models.Group
	if err := initializers.DB.Where("id = ?", deleteMessagesRequest.GroupID).First(&group).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Group not found"})
	}

	var messages []models.GroupMessage
	if err := initializers.DB.Where("id IN (?) AND group_id = ?", deleteMessagesRequest.MessageIDs, deleteMessagesRequest.GroupID).Find(&messages).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Messages not found"})
	}

	for _, message := range messages {
		if message.UserID.String() != user.ID.String() {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "You are not allowed to delete this message"})
		}
	}

	if err := initializers.DB.Delete(&messages).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to delete messages"})
	}

	return c.JSON(fiber.Map{"message": "Messages deleted successfully"})
}
