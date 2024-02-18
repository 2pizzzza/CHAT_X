package data_accsess

import (
	"github.com/google/uuid"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
)

func GetUserByID(userID uuid.UUID) (*models.User, error) {
	user := &models.User{}
	if err := initializers.DB.First(user, "id = ?", userID).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func GetMessageByID(messageID uint) (*models.Message, error) {
	message := &models.Message{}
	if err := initializers.DB.First(message, "id = ?", messageID).Error; err != nil {
		return nil, err
	}
	return message, nil
}
