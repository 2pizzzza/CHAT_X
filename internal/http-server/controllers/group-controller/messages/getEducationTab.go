package messages

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/auth-controllers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
)

func GetAllEducationGroups(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	user, err := auth_controllers.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	var groups []*models.Group
	if err := initializers.DB.Model(&models.Group{}).Where("target = 'education' AND id IN (SELECT group_id FROM group_participants WHERE user_id = ?)", user.ID).Find(&groups).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch education groups"})
	}

	PopulateLastMessage(groups)
	return c.JSON(groups)
}
func PopulateLastMessage(groups []*models.Group) {
	for _, group := range groups {
		var lastMessage models.GroupMessage
		if err := initializers.DB.Where("group_id = ?", group.ID).Order("created_at desc").First(&lastMessage).Error; err != nil {
			continue
		}
		group.LastMessage = &lastMessage
	}
}
