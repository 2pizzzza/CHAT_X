package group_controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/auth-controllers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
)

func CreateGroup(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	user, err := auth_controllers.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	var createGroupRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		PhotoURL    string `json:"photo_url"`
	}
	if err := c.BodyParser(&createGroupRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request"})
	}

	var existingGroup models.Group
	if initializers.DB.Where("name = ?", createGroupRequest.Name).First(&existingGroup).Error == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Group with this name already exists"})
	}

	newGroup := &models.Group{
		Name:         createGroupRequest.Name,
		Description:  createGroupRequest.Description,
		PhotoURL:     createGroupRequest.PhotoURL,
		Admins:       []*models.User{&user},
		Participants: []*models.User{&user},
	}

	if err := initializers.DB.Create(newGroup).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to create group"})
	}

	return c.JSON(fiber.Map{"message": "Group created successfully", "group": models.FilterGroupRecord(newGroup)})
}
