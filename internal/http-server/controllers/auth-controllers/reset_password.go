package auth_controllers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	initializers2 "github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
	"golang.org/x/crypto/bcrypt"
	"net/smtp"
	"time"
)

const resetCodeExpiry = time.Hour

func ResetPasswordRequest(c *fiber.Ctx) error {
	var request struct {
		Email string `json:"email"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Invalid request payload"})
	}

	var user models.User
	if err := initializers2.DB.Where("email = ?", request.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "fail", "message": "User not found"})
	}

	resetCode := generateVerificationCode()

	user.ResetPasswordCode = resetCode
	user.ResetPasswordExpiry = time.Now().Add(resetCodeExpiry)
	if err := initializers2.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to save reset code"})
	}

	err := sendResetCodeEmail(request.Email, resetCode)
	if err != nil {
		// В случае ошибки отправки письма можно обработать ее соответствующим образом
		fmt.Println("Error sending reset code email:", err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Reset code sent successfully"})
}

func ResetPasswordVerify(c *fiber.Ctx) error {
	var request struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Invalid request payload"})
	}

	var user models.User
	if err := initializers2.DB.Where("email = ?", request.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "fail", "message": "User not found"})
	}

	if user.ResetPasswordCode != request.Code || time.Now().After(user.ResetPasswordExpiry) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "Invalid or expired reset code"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Reset code verified successfully"})
}

func ResetPassword(c *fiber.Ctx) error {
	var request struct {
		Email    string `json:"email"`
		Code     string `json:"code"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Invalid request payload"})
	}

	var user models.User
	if err := initializers2.DB.Where("email = ?", request.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "fail", "message": "User not found"})
	}

	if user.ResetPasswordCode != request.Code || time.Now().After(user.ResetPasswordExpiry) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "Invalid or expired reset code"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to generate hashed password"})
	}

	user.Password = string(hashedPassword)
	user.ResetPasswordCode = ""
	user.ResetPasswordExpiry = time.Time{}
	if err := initializers2.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to update password"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Password updated successfully"})
}

func sendResetCodeEmail(email, code string) error {
	auth := smtp.PlainAuth("", "eligdigital@gmail.com", "dqwqqgtxbbuwobgt", "smtp.gmail.com")
	to := []string{email}

	htmlMsg := `
    <html>
    <body>
        <h1 style="text-align: center;">Email Verification Code</h1>
        <p style="text-align: center; font-size: 20px;">Your verification code is:</p>
        <div style="text-align: center; font-size: 30px; border: 2px solid #000; padding: 10px; margin: 20px;">
            <h3> ` + code + `<h3/>
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
