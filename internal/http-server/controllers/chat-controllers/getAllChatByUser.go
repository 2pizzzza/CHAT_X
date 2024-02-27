package chat_controllers

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/auth-controllers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
	"gorm.io/gorm"
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

	var responseChats []interface{}

	// Get personal chats
	var chats []models.Chat
	if result := initializers.DB.Where("user1_id = ? OR user2_id = ?", user.ID, user.ID).Find(&chats); result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch chats"})
	}

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
				Name:        user2.Name,
				User2ID:     *user2.ID,
				LastMessage: lastMessage.Text,
				CreatedAt:   chat.CreatedAt,
				UpdatedAt:   chat.UpdatedAt,
				Class:       "chat",
			}
			responseChats = append(responseChats, responseChat)
		} else {
			responseChat := models.ResponseChat{
				ID:          chat.ID,
				Name:        user1.Name,
				User1ID:     *user1.ID,
				User2ID:     *user2.ID,
				LastMessage: lastMessage.Text,
				CreatedAt:   chat.CreatedAt,
				UpdatedAt:   chat.UpdatedAt,
				Class:       "chat",
			}
			responseChats = append(responseChats, responseChat)
		}
	}

	// Get group chats
	var groups []models.Group
	if result := initializers.DB.Preload("Message").Where("id IN (SELECT group_id FROM group_participants WHERE user_id = ?) AND target = ?", user.ID, "personal").Find(&groups); result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch groups"})
	}

	for _, group := range groups {
		var lastMessage models.GroupMessage
		result := initializers.DB.Order("created_at desc").Where("group_id = ?", group.ID).First(&lastMessage)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				lastMessage.Text = "" // Set text to empty string if message not found
			} else {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch last message"})
			}
		}

		responseGroup := models.GroupResponse{
			ID:           group.ID,
			Name:         group.Name,
			Description:  group.Description,
			PhotoURL:     group.PhotoURL,
			Participants: group.Participants,
			LastMessage:  lastMessage.Text,
			CreatedAt:    group.CreatedAt,
			UpdatedAt:    group.UpdatedAt,
			Class:        "group",
		}
		responseChats = append(responseChats, responseGroup)
	}

	return c.JSON(responseChats)
}
