package main

import (
	"fmt"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/auth-controllers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/chat-controllers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/group-controller"
	"github.com/wpcodevo/golang-fiber-jwt/internal/http-server/controllers/group-controller/messages"
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
	app.Use(logger.New())
	app.Use(requestid.New())
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
	app.Get("/chat/signal/:chatID", func(c *fiber.Ctx) error {
		chatID := c.Params("chatID")
		chat_controllers.SignalChannels[chatID] <- true
		return c.SendStatus(fiber.StatusOK)
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
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	go chat_controllers.RunHub()

	micro.Route("/chat", func(router fiber.Router) {
		router.Post("/messages", middleware.DeserializeUser, chat_controllers.CreateMessage)
		router.Post("/delete-messages", middleware.DeserializeUser, chat_controllers.DeleteMessage)
		router.Get("/ws/:chatID", websocket.New(chat_controllers.HandlerWebSocketChat))
		router.Put("change-messages", middleware.DeserializeUser, chat_controllers.UpdateMessage)
		router.Post("/reply-messages", chat_controllers.ReplyToMessage)
		router.Get("/get-all-chat", chat_controllers.GetAllChatsByUser)
	})

	micro.Route("/group", func(router fiber.Router) {
		router.Post("/create-group", group_controller.CreateGroup)
		router.Post("/add-new-participant", group_controller.AddParticipantToGroup)
		router.Post("/delete-participant", group_controller.RemoveParticipant)
		router.Post("/added-new-admin", group_controller.AddAdmin)
		router.Post("/create-messages-group", messages.CreateGroupMessage)
		router.Post("/delete-messages", messages.DeleteGroupMessages)
		router.Put("/change-messages", messages.UpdateGroupMessage)
		router.Get("/get-all-group", group_controller.GetUserGroups)
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
