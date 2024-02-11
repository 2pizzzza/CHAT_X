package main

import (
	"fmt"
	controllers "github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/middleware"
	initializers "github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
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
		router.Post("/register", controllers.SignUpUser)
		router.Post("/login", controllers.SignInUser)
		router.Get("/logout", middleware.DeserializeUser, controllers.LogoutUser)
		router.Post("/confirm", controllers.ConfirmUser)
		router.Post("/refresh", controllers.RefreshAccessToken)
		router.Post("/change-password", controllers.ChangePassword)
		router.Get("/verify-email", controllers.VerifyEmail)
		router.Post("/reset-password-request", controllers.ResetPasswordRequest)
		router.Post("/reset-password-verify", controllers.ResetPasswordVerify)
		router.Post("/reset-password", controllers.ResetPassword)
	})

	micro.Get("/users/me", middleware.DeserializeUser, controllers.GetMe)

	micro.Get("/healthchecker", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": "JWT Authentication with Golang, Fiber, and GORM",
		})
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
