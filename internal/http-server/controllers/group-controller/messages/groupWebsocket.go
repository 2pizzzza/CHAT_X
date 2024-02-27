package messages

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
	"log/slog"
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
	var reactions []*models.Reaction

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
			message.Read = true
			initializers.DB.Save(&message)
			responseMessage := models.FilterGroupMessageRecord(&message)
			if err := initializers.DB.Model(&message).Association("Reactions").Find(&reactions); err != nil {
				log.Println("failed to load reactions for message:", err)
				return
			}
			responseMessage.Reaction = reactions
			if message.ParentMessageID != nil {
				var parentMessage models.GroupMessage
				if err := initializers.DB.Where("id = ?", *message.ParentMessageID).First(&parentMessage).Error; err != nil {
					slog.Error("failed to get parent message:", err)
				} else {
					responseMessage.ParentMessageID = &parentMessage.ID
					var parentUsername string
					if err := initializers.DB.Model(&parentMessage.User).Select("name").First(&parentUsername).Error; err != nil {
						slog.Error("failed to get parent Username:", err)
					}
					responseMessage.ParentMessage = &models.ParentMessagesGroup{
						ID:       parentMessage.ID,
						Username: parentUsername,
						Text:     parentMessage.Text,
					}
				}
			}
			var username string
			if err := initializers.DB.Model(&message.User).Select("name").First(&username).Error; err != nil {
				slog.Error("failed to get username:", err)
			}

			responseMessage.Username = username

			if err := c.WriteJSON(responseMessage); err != nil {
				slog.Error("failed, to send message:", err)
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

				}

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
				if err := initializers.DB.Model(&message).Association("Reactions").Find(&reactions); err != nil {
					log.Println("failed to load reactions for message:", err)
					return
				}

				responseMessage := models.FilterGroupMessageRecord(&message)
				responseMessage.Reaction = reactions
				if message.ParentMessageID != nil {
					var parentMessage models.GroupMessage
					if err := initializers.DB.Where("id = ?", *message.ParentMessageID).First(&parentMessage).Error; err != nil {
						slog.Error("failed to get parent message:", err)
					} else {
						responseMessage.ParentMessageID = &parentMessage.ID
						var parentUsername string
						if err := initializers.DB.Model(&parentMessage.User).Select("name").First(&parentUsername).Error; err != nil {
							slog.Error("failed to get parent Username:", err)
						}
						responseMessage.ParentMessage = &models.ParentMessagesGroup{
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

					message, err := AddReaction(userID, messageID, emoji)
					if err != nil {
						log.Println("failed to add reaction:", err)
						continue
					}
					updateMessage := models.FilterGroupMessageRecord(message)
					if err := initializers.DB.Model(&message).Association("Reactions").Find(&reactions); err != nil {
						log.Println("failed to load reactions for message:", err)
						return
					}
					updateMessage.Reaction = reactions
					if message.ParentMessageID != nil {
						var parentMessage models.GroupMessage
						if err := initializers.DB.Where("id = ?", *message.ParentMessageID).First(&parentMessage).Error; err != nil {
							slog.Error("failed to get parent message:", err)
						} else {
							updateMessage.ParentMessageID = &parentMessage.ID
							var parentUsername string
							if err := initializers.DB.Model(&parentMessage.User).Select("name").First(&parentUsername).Error; err != nil {
								slog.Error("failed to get parent Username:", err)
							}
							updateMessage.ParentMessage = &models.ParentMessagesGroup{
								ID:       parentMessage.ID,
								Username: parentUsername,
								Text:     parentMessage.Text,
							}
						}
					}
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

					updatedMessage := models.FilterGroupMessageRecord(message)
					if err := initializers.DB.Model(&message).Association("Reactions").Find(&reactions); err != nil {
						log.Println("failed to load reactions for message:", err)
						return
					}
					updatedMessage.Reaction = reactions
					if message.ParentMessageID != nil {
						var parentMessage models.GroupMessage
						if err := initializers.DB.Where("id = ?", *message.ParentMessageID).First(&parentMessage).Error; err != nil {
							slog.Error("failed to get parent message:", err)
						} else {
							updatedMessage.ParentMessageID = &parentMessage.ID
							var parentUsername string
							if err := initializers.DB.Model(&parentMessage.User).Select("name").First(&parentUsername).Error; err != nil {
								slog.Error("failed to get parent Username:", err)
							}
							updatedMessage.ParentMessage = &models.ParentMessagesGroup{
								ID:       parentMessage.ID,
								Username: parentUsername,
								Text:     parentMessage.Text,
							}
						}
					}

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

func AddReaction(userID, messageID string, emoji string) (*models.GroupMessage, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	messageIDUint, err := strconv.ParseUint(messageID, 10, 32)
	if err != nil {
		return nil, err
	}

	var existingReaction models.Reaction
	if err := initializers.DB.Where("user_id = ? AND message_id = ?", userUUID, messageIDUint).First(&existingReaction).Error; err != nil {
		// Если реакции от пользователя на это сообщение нет, создаем новую реакцию
		if errors.Is(err, gorm.ErrRecordNotFound) {
			reaction := models.Reaction{
				Emoji:     emoji,
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

	var message models.GroupMessage
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
		reaction := models.Reaction{
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

func RemoveReaction(userID, reactionID string) (*models.GroupMessage, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	reactionIDUint, err := strconv.ParseUint(reactionID, 10, 32)
	if err != nil {
		return nil, err
	}

	var reaction models.Reaction
	if err := initializers.DB.Where("id = ? AND user_id = ?", reactionIDUint, userUUID).First(&reaction).Error; err != nil {
		return nil, err
	}

	if err := initializers.DB.Exec("DELETE FROM group_message_reactions WHERE group_message_id = ? AND reaction_id = ?", reaction.MessageID, reaction.ID).Error; err != nil {
		return nil, err
	}

	if err := initializers.DB.Delete(&reaction).Error; err != nil {
		return nil, err
	}

	if err := initializers.DB.Delete(&models.Reaction{}, reactionID).Error; err != nil {
		return nil, err
	}

	var updatedMessage models.GroupMessage
	if err := initializers.DB.Where("id = ?", reaction.MessageID).Preload("Reactions").First(&updatedMessage).Error; err != nil {
		return nil, err
	}

	return &updatedMessage, nil
}
