package chat_controllers

import (
	"errors"
	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/utills/jwt_utils"
	"github.com/wpcodevo/golang-fiber-jwt/models"
	"gorm.io/gorm"
	"log"
	"strconv"
)

func HandlerWebSocketChat(c *websocket.Conn) {
	chatID := c.Params("chatID")
	claims, err := jwt_utils.ValidateToken(c)
	if err != nil {
		log.Println("failed to validate token:", err.Error())
		c.Close()
		return
	}

	userID := claims["sub"].(string)
	var user models.User
	if err := initializers.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		log.Println("failed to get user:", err)
		c.Close()
		return
	}

	defer func() {
		unregister <- c
		c.Close()
	}()

	register <- c

	log.Printf("WebSocket connection established for chat ID %s", chatID)

	var offset int
	const pageSize = 50

	var lastMessage models.Message
	if err := initializers.DB.Where("chat_id = ?", chatID).First(&lastMessage).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Println("no messages in this chat")
			return
		}
		log.Println("failed to get last message:", err)
		return
	}

	if err := initializers.DB.Where("chat_id = ?", chatID).Order("id desc").First(&lastMessage).Error; err != nil {
		log.Println("failed to get last message:", err)
		return
	}

	var messages []models.Message

	if err := initializers.DB.Where("chat_id = ?", chatID).Order("created_at desc").Limit(pageSize).Find(&messages).Error; err != nil {
		log.Println("failed to load messages:", err)
		return
	}

	for i := len(messages) - 1; i >= 0; i-- {
		message := messages[i]
		message.Read = true
		initializers.DB.Save(&message)
		responseMessage := models.FilterMessageRecord(&message)
		if err := c.WriteJSON(responseMessage); err != nil {
			log.Println("failed to send message:", err)
			continue
		}
	}

	SignalChannels[chatID] = make(chan bool)
	chatSignals[chatID] = SignalChannels[chatID]

	for {
		select {
		case <-SignalChannels[chatID]:
			offset += pageSize
			var prevMessages []models.Message
			if err := initializers.DB.Where("chat_id = ?", chatID).Order("created_at desc").Offset(offset).Limit(pageSize).Find(&prevMessages).Error; err != nil {
				log.Println("failed to load previous messages:", err)
				continue
			}

			for i := len(prevMessages) - 1; i >= 0; i-- {
				message := prevMessages[i]
				message.Read = true
				initializers.DB.Save(&message)
				responseMessage := models.FilterMessageRecord(&message)
				if err := c.WriteJSON(responseMessage); err != nil {
					log.Println("failed to send message:", err)
					continue
				}
			}
		default:
			var request map[string]interface{}
			if err := c.ReadJSON(&request); err != nil {
				log.Println("failed to read message:", err)
				continue
			}

			text, ok := request["text"].(string)
			if !ok {
				log.Println("invalid text format")
				continue
			}

			parentMessageID, ok := request["parentMessage"].(string)
			if !ok {
				log.Println("invalid parentMessage format", ok)
				continue
			}
			userUUID, err := uuid.Parse(userID)
			if err != nil {
				log.Println("failed to parse user UUID:", err)
				continue
			}
			chatIDInt, err := strconv.ParseUint(chatID, 10, 32)
			if err != nil {
				log.Println("failed to parse chat ID:", err)
				continue
			}
			parentMessageIDInt, err := strconv.ParseUint(parentMessageID, 10, 32)
			if err != nil {
				log.Println("failed to parse chat ID:", err)
				continue
			}
			parentMessageIDUint := uint(parentMessageIDInt)
			chatIDUint := uint(chatIDInt)

			message := models.Message{
				UserID:          &userUUID,
				ChatID:          chatIDUint,
				Text:            text,
				ParentMessageID: &parentMessageIDUint,
			}

			// Save the message to the database
			if err := initializers.DB.Create(&message).Error; err != nil {
				log.Println("failed to save message:", err)
				continue
			}

			// Send the new message to all connected clients
			responseMessage := models.FilterMessageRecord(&message)
			broadcast <- responseMessage
		}
	}
}

type client struct{}

type ChatSignal struct {
	ChatID string
	Signal chan bool
}

var SignalChannels = make(map[string]chan bool)
var chatSignals = make(map[string]chan bool)

var clients = make(map[*websocket.Conn]client)
var register = make(chan *websocket.Conn)
var broadcast = make(chan models.ResponseMessage)
var unregister = make(chan *websocket.Conn)

func RunHub() {
	for {
		select {
		case connection := <-register:
			clients[connection] = client{}
			log.Println("connection registered")

		case message := <-broadcast:
			log.Println("message received:", message)

			for connection := range clients {
				if err := connection.WriteJSON(message); err != nil {
					log.Println("write error:", err)

					unregister <- connection
					connection.WriteMessage(websocket.CloseMessage, []byte{})
					connection.Close()
				}
			}

		case connection := <-unregister:
			delete(clients, connection)
			log.Println("connection unregistered")
		}
	}
}
