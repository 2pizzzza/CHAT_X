package group_controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/auth-controllers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
	"path/filepath"
	"strconv"
)

func UpdateGroup(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	user, err := auth_controllers.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	groupID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid group ID"})
	}

	var group models.Group
	if err := initializers.DB.Preload("Admins").Where("id = ?", groupID).First(&group).Error; err != nil {
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
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "You are not an admin of this group"})
	}

	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	files := form.File["photo_url"]
	for _, file := range files {
		filename := filepath.Join("images", file.Filename)

		if err := c.SaveFile(file, filename); err != nil {
			return err
		}

		updateData := map[string]interface{}{
			"PhotoURL": filename,
		}
		if err := initializers.DB.Model(&group).Where("id = ?", groupID).Updates(updateData).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to update group"})
		}
	}

	return c.JSON(fiber.Map{"message": "Group updated successfully"})
}
