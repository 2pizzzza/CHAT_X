package chat_controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/auth-controllers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
)

func GetAllChatsByUser(c *fiber.Ctx) error {
	// Проверка аутентификации
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	user, err := auth_controllers.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	// Получение всех чатов пользователя
	var chats []models.Chat
	if result := initializers.DB.Where("user1_id = ? OR user2_id = ?", user.ID, user.ID).Find(&chats); result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch chats"})
	}

	// Формирование ответа
	var responseChats []models.ResponseChat
	for _, chat := range chats {
		responseChat := models.FilterChatRecord(&chat)
		responseChats = append(responseChats, responseChat)
	}

	return c.JSON(responseChats)
}
