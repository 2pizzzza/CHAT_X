package main

import (
	"fmt"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/auth-controllers"
	chat_controllers "github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/chat-controllers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/middleware"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"log"
)

func init() {
	config, err := initializers.LoadConfig(".")
	if err != nil {
		log.Fatalln("Failed to load environment variables! \n", err.Error())
	}
	initializers.ConnectDB(&config)
}

func main() {
	app := fiber.New()
	micro := fiber.New()

	app.Mount("/api", micro)
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowMethods:     "GET, POST",
		AllowCredentials: true,
	}))

	micro.Route("/auth", func(router fiber.Router) {
		router.Post("/register", auth_controllers.SignUpUser)
		router.Post("/login", auth_controllers.SignInUser)
		router.Get("/logout", middleware.DeserializeUser, auth_controllers.LogoutUser)
		router.Post("/confirm", auth_controllers.ConfirmUser)
		router.Post("/refresh", auth_controllers.RefreshAccessToken)
		router.Post("/change-password", auth_controllers.ChangePassword)
		router.Get("/verify-email", auth_controllers.VerifyEmail)
		router.Post("/reset-password-request", auth_controllers.ResetPasswordRequest)
		router.Post("/reset-password-verify", auth_controllers.ResetPasswordVerify)
		router.Post("/reset-password", auth_controllers.ResetPassword)
	})

	micro.Get("/users/me", middleware.DeserializeUser, auth_controllers.GetMe)

	micro.Get("/healthchecker", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": "JWT Authentication with Golang, Fiber, and GORM",
		})
	})
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	micro.Route("/chat", func(router fiber.Router) {
		router.Post("/messages", middleware.DeserializeUser, chat_controllers.CreateMessage)

	})

	micro.All("*", func(c *fiber.Ctx) error {
		path := c.Path()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "fail",
			"message": fmt.Sprintf("Path: %v does not exists on this server", path),
		})
	})

	log.Fatal(app.Listen(":8000"))
}
