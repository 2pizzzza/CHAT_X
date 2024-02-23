package sticker

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
	"gorm.io/gorm"
	"mime/multipart"
	"path/filepath"
)

type StickerUpload struct {
	PackName string     `form:"pack_name" binding:"required"`
	Stickers []*Sticker `form:"stickers" binding:"required"`
}

type Sticker struct {
	Image *multipart.FileHeader `form:"image" binding:"required"`
}

func UploadStickers(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	packName := form.Value["pack_name"][0]
	stickers := form.File["stickers"]

	// Проверяем, существует ли уже пак стикеров с таким именем
	var existingPack models.StickerPack
	if err := initializers.DB.Where("name = ?", packName).First(&existingPack).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if existingPack.ID != 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Sticker pack with this name already exists"})
	}

	// Создаем новый пак стикеров
	pack := models.StickerPack{Name: packName}
	if err := initializers.DB.Create(&pack).Error; err != nil {
		return err
	}

	// Сохраняем изображения стикеров в директорию
	for _, sticker := range stickers {
		file, err := sticker.Open()
		if err != nil {
			return err
		}
		defer file.Close()

		// Сохраняем файл в директорию images
		filename := filepath.Join("images", sticker.Filename)
		if err := c.SaveFile(sticker, filename); err != nil {
			return err
		}

		// Создаем запись о стикере в базе данных
		stickerDB := models.Sticker{PackID: pack.ID, URL: filename}
		if err := initializers.DB.Create(&stickerDB).Error; err != nil {
			return err
		}
	}

	return c.JSON(fiber.Map{"message": "Stickers uploaded successfully"})
}
