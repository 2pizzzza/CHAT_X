package messages

import (
	"fmt"
	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/internal/utills/jwt_utils"
	"github.com/wpcodevo/golang-fiber-jwt/models"
	"log"
	"strconv"
)

func HandlerWebSocketGroupMessages(c *websocket.Conn) {
	groupID := c.Params("GroupID")
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

	log.Printf("WebSocket connection established for group ID %s", groupID)

	var offset int
	const pageSize = 150

	var count int64
	if err := initializers.DB.Model(&models.GroupMessage{}).Where("group_id = ?", groupID).Count(&count).Error; err != nil {
		log.Println("failed to count messages:", err)
		return
	}
	if count > 0 {
		var lastMessage models.GroupMessage
		if err := initializers.DB.Where("group_id = ?", groupID).Order("id desc").First(&lastMessage).Error; err != nil {
			log.Println("failed to get last message:", err)
			return
		}

		var messages []models.GroupMessage
		if err := initializers.DB.Where("group_id = ?", groupID).Order("created_at desc").Limit(pageSize).Find(&messages).Error; err != nil {
			log.Println("failed to load messages:", err)
			return
		}

		for i := len(messages) - 1; i >= 0; i-- {
			message := messages[i]
			responseMessage := message
			message.Read = true
			initializers.DB.Save(&message)
			if err := c.WriteJSON(responseMessage); err != nil {
				log.Println("failed to send message:", err)
				continue
			}
		}
	}

	SignalChannels[groupID] = make(chan bool)
	groupSignal[groupID] = SignalChannels[groupID]

	for {
		select {
		case <-SignalChannels[groupID]:
			offset += pageSize
			var prevMessages []models.GroupMessage
			if err := initializers.DB.Where("group_id = ?", groupID).Order("created_at desc").Offset(offset).Limit(pageSize).Find(&prevMessages).Error; err != nil {
				log.Println("failed to load previous messages:", err)
				continue
			}

			for i := len(prevMessages) - 1; i >= 0; i-- {
				message := prevMessages[i]
				responseMessage := models.FilterGroupMessageRecord(&message)
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

			var parentMessageIDUint uint
			if parentMessage, ok := request["parentMessage"].(string); ok && parentMessage != "" {
				parentMessageUint, err := strconv.ParseUint(parentMessage, 10, 32)
				if err != nil {
					log.Println("invalid parentMessage format", err)
					continue
				}
				parentMessageIDUint = uint(parentMessageUint)
			}
			fmt.Println(parentMessageIDUint)

			userUUID, err := uuid.Parse(userID)
			if err != nil {
				log.Println("failed to parse user UUID:", err)
				continue
			}
			chatIDInt, err := strconv.ParseUint(groupID, 10, 32)
			if err != nil {
				log.Println("failed to parse chat ID:", err)
				continue
			}

			chatIDUint := uint(chatIDInt)
			var message models.GroupMessage

			if parentMessageIDUint > 0 {
				message = models.GroupMessage{
					UserID:          &userUUID,
					GroupID:         chatIDUint,
					Text:            text,
					ParentMessageID: &parentMessageIDUint,
				}
			} else {
				message = models.GroupMessage{
					UserID:  &userUUID,
					GroupID: chatIDUint,
					Text:    text,
				}
			}

			if err := initializers.DB.Create(&message).Error; err != nil {
				log.Println("failed to save message:", err)
				continue
			}

			responseMessage := models.FilterGroupMessageRecord(&message)
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
var groupSignal = make(map[string]chan bool)

var clients = make(map[*websocket.Conn]client)
var register = make(chan *websocket.Conn)
var broadcast = make(chan models.GroupMessageResponse)
var unregister = make(chan *websocket.Conn)

func RunHubGroup() {
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
