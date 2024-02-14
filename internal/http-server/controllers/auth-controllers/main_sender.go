package auth_controllers

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
	"strings"

	"net/smtp"
)

// link
func VerifyEmail(c *fiber.Ctx) error {
	confirmationLink := c.Query("link")

	var user models.User
	if err := initializers.DB.Where("confirmation_link = ?", confirmationLink).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "fail", "message": "Invalid confirmation link"})
	}

	user.Verified = true
	if err := initializers.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to verify email"})
	}

	return c.Status(fiber.StatusOK).SendFile("../../template/verifie.html")
}

func ConfirmUser(c *fiber.Ctx) error {
	var confirmation struct {
		Code string `json:"code"`
	}
	if err := c.BodyParser(&confirmation); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	user, err := GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "Unauthorized"})
	}

	if user.ConfirmationCode != confirmation.Code {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "fail", "message": "Invalid confirmation code"})
	}

	user.Verified = true
	user.ConfirmationCode = ""

	if err := initializers.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to update user"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "User confirmed successfully"})
}

func GetUserFromToken(c *fiber.Ctx) (models.User, error) {
	token := c.Get("Authorization")
	if token == "" {
		return models.User{}, errors.New("missing token")
	}

	token = strings.ReplaceAll(token, "Bearer ", "")
	claims := jwt.MapClaims{}

	config, err := initializers.LoadConfig(".")
	if err != nil {
		return models.User{}, err
	}

	_, err = jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JwtSecret), nil
	})
	if err != nil {
		return models.User{}, err
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return models.User{}, errors.New("invalid token")
	}

	var user models.User
	if err := initializers.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return models.User{}, err
	}

	return user, nil
}
func sendVerificationEmail(email, code string) error {
	auth := smtp.PlainAuth("", "eligdigital@gmail.com", "dqwqqgtxbbuwobgt", "smtp.gmail.com")
	to := []string{email}

	htmlMsg := `
    <html>
    <body>
        <h1 style="text-align: center;">Email Verification Code</h1>
        <p style="text-align: center; font-size: 20px;">Your verification code is:</p>
        <div style="text-align: center; font-size: 30px; border: 2px solid #000; padding: 10px; margin: 20px;">
            <a href="http://localhost:8000/api/auth-controllers/verify-email?link=` + code + `">кодддд<a/>
        </div>
    </body>
    </html>
    `

	err := smtp.SendMail("smtp.gmail.com:587", auth, "eligdigital@gmail.com", to, []byte(
		"From: eligdigital@gmail.com\r\n"+
			"To: "+email+"\r\n"+
			"Subject: Email Verification Code\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=utf-8\r\n"+
			"\r\n"+
			htmlMsg,
	))
	if err != nil {
		return err
	}
	return nil
}
func generateVerificationCode() string {
	randomBytes := make([]byte, 6)
	rand.Read(randomBytes)
	return base64.URLEncoding.EncodeToString(randomBytes)
}
