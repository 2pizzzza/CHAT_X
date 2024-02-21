package messages

import (
	"github.com/gofiber/fiber/v2"
	auth_controllers "github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/auth-controllers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
)

func ReplyToGroupMessage(c *fiber.Ctx) error {
	// Проверяем авторизацию пользователя
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	user, err := auth_controllers.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	// Парсим данные из запроса
	var replyMessageRequest struct {
		GroupID         uint   `json:"group_id"`
		ParentMessageID uint   `json:"parent_message_id"`
		Text            string `json:"text"`
	}
	if err := c.BodyParser(&replyMessageRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request"})
	}

	// Проверяем, что сообщение, на которое отвечаем, существует и принадлежит указанной группе
	var parentMessage models.GroupMessage
	if err := initializers.DB.Where("id = ? AND group_id = ?", replyMessageRequest.ParentMessageID, replyMessageRequest.GroupID).First(&parentMessage).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Parent message not found"})
	}

	// Создаем новое сообщение в группе в качестве ответа
	reply := &models.GroupMessage{
		GroupID:         replyMessageRequest.GroupID,
		UserID:          user.ID,
		ParentMessageID: &replyMessageRequest.ParentMessageID,
		Text:            replyMessageRequest.Text,
	}

	if err := initializers.DB.Create(reply).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to create reply message"})
	}

	// Возвращаем успешный ответ
	return c.JSON(fiber.Map{"message": "Reply message created successfully", "messages": models.FilterGroupMessageRecord(reply), "user": models.FilterUserRecord(&user), "parent": models.FilterGroupMessageRecord(&parentMessage)})
}
