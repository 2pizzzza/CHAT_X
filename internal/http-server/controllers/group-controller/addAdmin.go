package group_controller

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/auth-controllers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
)

func AddAdmin(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	user, err := auth_controllers.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	var addAdminRequest struct {
		GroupID uint      `json:"group_id"`
		UserID  uuid.UUID `json:"user_id"`
	}
	if err := c.BodyParser(&addAdminRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request"})
	}

	var group models.Group
	if err := initializers.DB.Preload("Admins").Where("id = ?", addAdminRequest.GroupID).First(&group).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Group not found"})
	}

	if err := initializers.DB.Preload("Participants").Where("id = ?", addAdminRequest.GroupID).First(&group).Error; err != nil {
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
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Only admins can add new admins"})
	}

	var isParticipant bool
	for _, participant := range group.Participants {
		fmt.Println(participant.ID.String())
		if participant.ID.String() == addAdminRequest.UserID.String() {
			fmt.Println(addAdminRequest.UserID.String(), participant.ID.String())
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "User is not a participant in the group"})
	}

	if err := initializers.DB.Model(&group).Association("Admins").Append(&models.User{ID: &addAdminRequest.UserID}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to add admin to group"})
	}

	// Возвращаем успешный ответ
	return c.JSON(fiber.Map{"message": "Admin added successfully"})
}
