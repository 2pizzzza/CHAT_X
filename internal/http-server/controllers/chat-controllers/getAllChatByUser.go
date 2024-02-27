package chat_controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/auth-controllers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
)

func GetAllChatsByUser(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	user, err := auth_controllers.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	var chats []models.Chat
	if result := initializers.DB.Where("user1_id = ? OR user2_id = ?", user.ID, user.ID).Find(&chats); result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch chats"})
	}

	var responseChats []models.ResponseChat
	for _, chat := range chats {
		var user1, user2 models.User
		if err := initializers.DB.First(&user1, chat.User1ID).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch user1"})
		}
		if err := initializers.DB.First(&user2, chat.User2ID).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch user2"})
		}

		var lastMessage models.Message
		if result := initializers.DB.Order("created_at desc").Where("chat_id = ?", chat.ID).First(&lastMessage); result.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch last message"})
		}
		if user.ID == user1.ID {
			responseChat := models.ResponseChat{
				ID:          chat.ID,
				User2Name:   user1.Name,
				User1ID:     *user1.ID,
				User2ID:     *user2.ID,
				LastMessage: lastMessage.Text,
				CreatedAt:   chat.CreatedAt,
				UpdatedAt:   chat.UpdatedAt,
			}
			responseChats = append(responseChats, responseChat)
		} else {
			responseChat := models.ResponseChat{
				ID:          chat.ID,
				User1Name:   user2.Name,
				User1ID:     *user1.ID,
				User2ID:     *user2.ID,
				LastMessage: lastMessage.Text,
				CreatedAt:   chat.CreatedAt,
				UpdatedAt:   chat.UpdatedAt,
			}
			responseChats = append(responseChats, responseChat)
		}

	}

	return c.JSON(responseChats)
}
