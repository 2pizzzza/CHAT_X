package data_accsess

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/auth-controllers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
)

func GetGroupsByTarget(c *fiber.Ctx) error {
	user, err := auth_controllers.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	target := c.Params("target")

	var groups []*models.Group
	if err := initializers.DB.Model(&user).Where("target = ? AND id IN (SELECT group_id FROM group_participants WHERE user_id = ?)", target, user.ID).Find(&groups).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch groups"})
	}

	return c.JSON(groups)
}
