package messages

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/auth-controllers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
)

func CreateGroupMessage(c *fiber.Ctx) error {
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
	var createMessageRequest struct {
		GroupID uint   `json:"group_id"`
		Text    string `json:"text"`
	}
	if err := c.BodyParser(&createMessageRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request"})
	}

	// Проверяем, является ли текущий пользователь участником группы
	var group models.Group
	if err := initializers.DB.Preload("Participants").Where("id = ?", createMessageRequest.GroupID).First(&group).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Group not found"})
	}

	var isParticipant bool
	for _, participant := range group.Participants {
		if participant.ID.String() == user.ID.String() {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Only group participants can create messages"})
	}

	// Создаем новое сообщение в группе
	newMessage := &models.GroupMessage{
		GroupID: createMessageRequest.GroupID,
		UserID:  user.ID,
		Text:    createMessageRequest.Text,
	}

	if err := initializers.DB.Create(newMessage).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to create message"})
	}

	// Возвращаем успешный ответ
	return c.JSON(fiber.Map{"messages": "Message created successfully", "message": models.FilterGroupMessageRecord(newMessage)})
}
