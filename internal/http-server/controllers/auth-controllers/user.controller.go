package auth_controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
	"path/filepath"
)

func GetMe(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserResponse)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": fiber.Map{"user": user}})
}

type UserProfileUpdateRequest struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Photo  string `json:"photo"`
	Online bool   `json:"online"`
}

func UpdateUserProfile(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	user, err := GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}
	if name := form.Value["name"]; len(name) > 0 {
		user.Name = name[0]
	}
	if email := form.Value["email"]; len(email) > 0 {
		user.Email = email[0]
	}

	files := form.File["photo_url"]
	for _, file := range files {
		filename := filepath.Join("images", file.Filename)

		if err := c.SaveFile(file, filename); err != nil {
			return err
		}
		user.Photo = &filename
	}

	if err := initializers.DB.Save(&user).Error; err != nil {
		return err
	}

	return c.JSON(user)
}
