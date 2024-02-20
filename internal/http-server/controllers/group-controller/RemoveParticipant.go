package group_controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/auth-controllers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
)

func RemoveParticipant(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	user, err := auth_controllers.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	var removeParticipantRequest struct {
		GroupID uint      `json:"group_id"`
		UserID  uuid.UUID `json:"user_id"`
	}
	if err := c.BodyParser(&removeParticipantRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request"})
	}

	var group models.Group
	if err := initializers.DB.Preload("Admins").Preload("Participants").Where("id = ?", removeParticipantRequest.GroupID).First(&group).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Group not found"})
	}

	isAdmin := false
	for _, admin := range group.Admins {
		if admin.ID.String() == user.ID.String() {
			isAdmin = true
			break
		}
	}

	if user.ID != &removeParticipantRequest.UserID && !isAdmin {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "You are not allowed to remove this participant"})
	}

	var isParticipant bool
	for _, participant := range group.Participants {
		if participant.ID == &removeParticipantRequest.UserID {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "User is not a participant in the group"})
	}

	if err := initializers.DB.Model(&group).Association("Participants").Delete(&models.User{ID: &removeParticipantRequest.UserID}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to remove participant from group"})
	}

	return c.JSON(fiber.Map{"message": "Participant removed successfully"})
}
