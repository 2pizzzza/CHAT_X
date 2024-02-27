package chat_controllers

import (
	"errors"
	"fmt"
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

		if message.ParentMessageID != nil {
			var parentMessage models.Message
			if err := initializers.DB.Where("id = ?", *message.ParentMessageID).First(&parentMessage).Error; err != nil {
				log.Println("failed to get parent message:", err)
			} else {
				responseMessage.ParentMessageID = &parentMessage.ID
				var parentUsername string
				if err := initializers.DB.Model(&parentMessage.User).Select("name").First(&parentUsername).Error; err != nil {
					log.Println("failed to get parent username:", err)
				}
				responseMessage.ParentMessage = &models.ParentMessages{
					ID:       parentMessage.ID,
					Username: parentUsername,
					Text:     parentMessage.Text,
				}
			}
		}

		var username string
		if err := initializers.DB.Model(&message.User).Select("name").First(&username).Error; err != nil {
			log.Println("failed to get username:", err)
		}
		responseMessage.Username = username

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

			messageType, ok := request["type"].(string)
			if !ok {
				log.Println("invalid message type")
				continue
			}

			switch messageType {
			case "message":
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
					fmt.Println(parentMessageIDUint)
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

				chatIDUint := uint(chatIDInt)
				var message models.Message

				if parentMessageIDUint > 0 {
					message = models.Message{
						UserID:          &userUUID,
						ChatID:          chatIDUint,
						Text:            text,
						ParentMessageID: &parentMessageIDUint,
					}
				} else {
					message = models.Message{
						UserID: &userUUID,
						ChatID: chatIDUint,
						Text:   text,
					}
				}

				if parentMessageIDUint > 0 {
					message.ParentMessageID = &parentMessageIDUint
				}
				fmt.Println(message)
				if err := initializers.DB.Create(&message).Error; err != nil {
					log.Println("failed to save message:", err)
					continue
				}

				responseMessage := models.FilterMessageRecord(&message)
				if message.ParentMessageID != nil {
					var parentMessage models.Message
					if err := initializers.DB.Where("id = ?", *message.ParentMessageID).First(&parentMessage).Error; err != nil {
						log.Println("failed to get parent message:", err)
					} else {
						responseMessage.ParentMessageID = &parentMessage.ID
						var parentUsername string
						if err := initializers.DB.Model(&parentMessage.User).Select("name").First(&parentUsername).Error; err != nil {
							log.Println("failed to get parent username:", err)
						}
						responseMessage.ParentMessage = &models.ParentMessages{
							ID:       parentMessage.ID,
							Username: parentUsername,
							Text:     parentMessage.Text,
						}
					}
				}
				broadcast <- responseMessage

			case "reaction":
				operation, ok := request["operation"].(string)
				if !ok {
					log.Println("invalid operation format")
					continue
				}
				switch operation {
				case "add":
					emoji, ok := request["emoji"].(string)
					if !ok {
						fmt.Println(emoji)
						log.Println("invalid emoji format")
						continue
					}
					messageID, ok := request["messageID"].(string)
					if !ok {
						log.Println("invalid messageID format")
						continue
					}

					message, err := AddReaction(userID, user.Name, messageID, emoji)
					if err != nil {
						log.Println("failed to add reaction:", err)
						continue
					}
					updateMessage := models.FilterMessageRecord(message)
					broadcast <- updateMessage

				case "remove":
					reactionID, ok := request["reactionID"].(string)
					if !ok {
						log.Println("invalid reactionID format")
						continue
					}

					message, err := RemoveReaction(userID, reactionID)
					if err != nil {
						log.Println("failed to remove reaction:", err, message)
						continue
					}

					updatedMessage := models.FilterMessageRecord(message)
					broadcast <- updatedMessage

				default:
					log.Println("unsupported operation:", operation)
				}
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

func AddReaction(userID, messageID, username string, emoji string) (*models.Message, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	messageIDUint, err := strconv.ParseUint(messageID, 10, 32)
	if err != nil {
		return nil, err
	}

	var existingReaction models.ChatReaction
	if err := initializers.DB.Where("user_id = ? AND message_id = ?", userUUID, messageIDUint).First(&existingReaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			reaction := models.Reaction{
				Emoji:     emoji,
				Username:  username,
				UserID:    &userUUID,
				MessageID: uint(messageIDUint),
			}
			if err := initializers.DB.Create(&reaction).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		existingReaction.Emoji = emoji
		if err := initializers.DB.Save(&existingReaction).Error; err != nil {
			return nil, err
		}
	}

	var message models.Message
	if err := initializers.DB.Where("id = ?", messageIDUint).Preload("Reactions").First(&message).Error; err != nil {
		return nil, err
	}

	reactionExists := false
	for _, reaction := range message.Reactions {
		if reaction.UserID.String() == userID {
			reaction.Emoji = emoji
			reactionExists = true
			break
		}
	}

	if !reactionExists {
		reaction := models.ChatReaction{
			Emoji:     emoji,
			UserID:    &userUUID,
			MessageID: uint(messageIDUint),
		}
		message.Reactions = append(message.Reactions, &reaction)
	}

	if err := initializers.DB.Save(&message).Error; err != nil {
		return nil, err
	}

	return &message, nil
}
func RemoveReaction(userID, reactionID string) (*models.Message, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	reactionIDUint, err := strconv.ParseUint(reactionID, 10, 32)
	if err != nil {
		return nil, err
	}

	var reaction models.ChatReaction
	if err := initializers.DB.Where("id = ? AND user_id = ?", reactionIDUint, userUUID).First(&reaction).Error; err != nil {
		return nil, err
	}

	if err := initializers.DB.Delete(&reaction).Error; err != nil {
		return nil, err
	}

	if err := initializers.DB.Delete(&models.ChatReaction{}, reactionID).Error; err != nil {
		return nil, err
	}

	var updatedMessage models.Message
	if err := initializers.DB.Where("id = ?", reaction.MessageID).Preload("Reactions").First(&updatedMessage).Error; err != nil {
		return nil, err
	}

	return &updatedMessage, nil
}
