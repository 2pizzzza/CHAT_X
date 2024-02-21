package messages

import (
	"fmt"
	"github.com/gofiber/contrib/websocket"
	"github.com/wpcodevo/golang-fiber-jwt/internal/storage/initializers"
	"github.com/wpcodevo/golang-fiber-jwt/models"
	"log"
	"time"
)

func HandlerWebSocketGroupMessages(c *websocket.Conn) {
	groupID := c.Params("GroupId")

	defer func() {
		unregister <- c
		c.Close()
	}()

	register <- c
	fmt.Println(groupID)

	log.Printf("WebSocket connection established for group ID %s", groupID)

	var lastMessageID uint
	var offset int
	const pageSize = 50

	var lastMessage models.GroupMessage
	if err := initializers.DB.Where("group_id = ?", groupID).Order("id desc").First(&lastMessage).Error; err != nil {
		log.Println("failed to get last message:", err)
		return
	}
	lastMessageID = lastMessage.ID

	var messages []models.GroupMessage
	if err := initializers.DB.Where("group_id = ?", groupID).Order("created_at desc").Limit(pageSize).Find(&messages).Error; err != nil {
		log.Println("failed to load messages:", err)
		return
	}

	for i := len(messages) - 1; i >= 0; i-- {
		message := messages[i]
		responseMessage := models.FilterGroupMessageRecord(&message)
		if err := c.WriteJSON(responseMessage); err != nil {
			log.Println("failed to send message:", err)
			continue
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
			for {
				var newMessages []models.GroupMessage
				if err := initializers.DB.Where("group_id = ? AND id > ?", groupID, lastMessageID).Order("id asc").Limit(pageSize).Find(&newMessages).Error; err != nil {
					log.Println("failed to load new messages:", err)
					continue
				}

				for _, message := range newMessages {
					responseMessage := models.FilterGroupMessageRecord(&message)
					if err := c.WriteJSON(responseMessage); err != nil {
						log.Println("failed to send new message:", err)
						continue
					}
				}

				if len(newMessages) > 0 {
					lastMessageID = newMessages[len(newMessages)-1].ID
				}

				time.Sleep(1 * time.Second)
			}
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
var broadcast = make(chan models.ResponseMessage)
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
