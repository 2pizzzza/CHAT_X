package chat_controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/auth-controllers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
)

func SearchGroupByName(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	user, err := auth_controllers.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}
	_ = user
	searchText := c.Query("text")
	if searchText == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Missing search text"})
	}

	var groups []models.Group
	result := initializers.DB.Where("name LIKE ?", "%"+searchText+"%").Find(&groups)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to search groups", "error": result.Error.Error()})
	}

	return c.JSON(groups)
}

type ChatSearchRequest struct {
	Text string `json:"text"`
}
