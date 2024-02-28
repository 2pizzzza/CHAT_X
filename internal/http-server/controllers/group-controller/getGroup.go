package group_controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
	"strconv"
)

func GetGroupInfo(c *fiber.Ctx) error {
	groupID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid group ID"})
	}

	group := models.Group{}
	result := initializers.DB.Preload("Message").Where("id = ?", groupID).First(&group)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Group not found"})
	}

	return c.JSON(group)
}
