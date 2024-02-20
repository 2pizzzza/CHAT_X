package group_controller

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/auth-controllers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
)

func AddParticipantToGroup(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	user, err := auth_controllers.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	var addParticipantRequest struct {
		GroupID uint       `json:"group_id"`
		UserID  *uuid.UUID `json:"user_id"`
	}
	if err := c.BodyParser(&addParticipantRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request"})
	}

	var group models.Group
	if err := initializers.DB.Preload("Admins").Where("id = ?", addParticipantRequest.GroupID).First(&group).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Group not found"})
	}

	isAdmin := false
	for _, admin := range group.Admins {
		if admin.ID.String() == user.ID.String() {
			isAdmin = true
			break
		}
	}
	if !isAdmin {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Only admins can add participants"})
	}

	var existingParticipant models.User
	if err := initializers.DB.Model(&group).Association("Participants").Find(existingParticipant.ID, "id = ?", addParticipantRequest.UserID); err == nil {
		fmt.Println(existingParticipant.ID, addParticipantRequest.UserID)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "User is already a participant in the group"})
	}

	if err := initializers.DB.Model(&group).Association("Participants").Append(&models.User{ID: addParticipantRequest.UserID}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to add participant to group"})
	}

	return c.JSON(fiber.Map{"message": "Participant added successfully"})
}
