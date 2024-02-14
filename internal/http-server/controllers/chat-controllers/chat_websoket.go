package chat_controllers

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	auth_controllers "github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/auth-controllers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
	"gorm.io/gorm"
)

type CreateMessageInput struct {
	RecipientID     uuid.UUID `json:"recipient_id" validate:"required"`
	Text            string    `json:"text" validate:"required"`
	ParentMessageID *uint     `json:"parent_message_id,omitempty"`
}

func CreateMessage(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	user, err := auth_controllers.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	var reqBody struct {
		RecipientID *uuid.UUID `json:"recipient_id"`
		Text        string     `json:"text"`
	}
	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}

	chat := &models.Chat{}
	result := initializers.DB.Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)",
		user.ID, reqBody.RecipientID, reqBody.RecipientID, user.ID).First(&chat)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		newChat := &models.Chat{
			User1ID: user.ID,
			User2ID: reqBody.RecipientID,
		}
		result := initializers.DB.Create(&newChat)
		if result.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error creating chat"})
		}
		chat = newChat
	} else if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error finding chat"})
	}

	message := &models.Message{
		UserID: user.ID,
		ChatID: chat.ID,
		Text:   reqBody.Text,
	}
	result = initializers.DB.Create(&message)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error creating message"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Message created successfully", "data": message})
}
