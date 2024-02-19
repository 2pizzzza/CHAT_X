package chat_controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/auth-controllers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
)

func ReplyToMessage(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	// Получаем пользователя из токена
	user, err := auth_controllers.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}
	// Парсинг параметров запроса
	var params struct {
		ChatID          uint   `json:"chat_id"`
		ParentMessageID uint   `json:"parent_message_id"`
		Text            string `json:"text"`
	}
	if err := c.BodyParser(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request payload"})
	}

	chat := &models.Chat{}
	if result := initializers.DB.Where("id = ? AND (user1_id = ? OR user2_id = ?)", params.ChatID, user.ID, user.ID).First(chat); result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch chat"})
	}

	if chat.ID == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Chat not found"})
	}

	message := &models.Message{
		UserID:          user.ID,
		ChatID:          params.ChatID,
		ParentMessageID: &params.ParentMessageID,
		Text:            params.Text,
	}
	if err := initializers.DB.Create(&message).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to create message"})
	}

	return c.JSON(fiber.Map{"message": "Message replied successfully"})
}
