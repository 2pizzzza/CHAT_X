package models

import (
	"github.com/google/uuid"
	"time"
)

type Chat struct {
	ID        uint       `gorm:"primaryKey"`
	User1     *User      `gorm:"foreignKey:User1ID"`
	User1ID   *uuid.UUID `gorm:"type:uuid"`
	User2     *User      `gorm:"foreignKey:User2ID"`
	User2ID   *uuid.UUID `gorm:"type:uuid"`
	Messages  []*Message `gorm:"foreignKey:ChatID"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
}

type Message struct {
	ID              uint            `gorm:"primaryKey"`
	User            *User           `gorm:"foreignKey:UserID"`
	UserID          *uuid.UUID      `gorm:"type:uuid"`
	Chat            Chat            `gorm:"foreignKey:ChatID"`
	ChatID          uint            `gorm:"not null"`
	ParentMessage   *ParentMessages `gorm:"-"`
	ParentMessageID *uint
	StickerID       *uint
	Sticker         Sticker         `gorm:"foreignKey:StickerID"`
	Read            bool            `gorm:"default:false"`
	Text            string          `gorm:"not null"`
	Reactions       []*ChatReaction `gorm:"many2many:chat_message_reactions;"`
	CreatedAt       time.Time       `gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime"`
}

type ChatReaction struct {
	ID        uint       `gorm:"primaryKey"`
	Emoji     string     `gorm:"not null"`
	UserID    *uuid.UUID `gorm:"type:uuid;not null"`
	MessageID uint       `gorm:"not null"`
	Username  string     `json:"username,omitempty"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
}

type ResponseMessage struct {
	ID              uint            `json:"id,omitempty"`
	UserID          uuid.UUID       `json:"user_id,omitempty"`
	User            *User           `json:"user,omitempty"`
	ChatID          uint            `json:"chat_id,omitempty"`
	ParentMessageID *uint           `json:"parent_message_id,omitempty"`
	ParentMessage   *ParentMessages `json:"parent_message,omitempty"`
	Text            string          `json:"text,omitempty"`
	Reaction        []*ChatReaction `json:"reaction,omitempty"`
	Read            bool            `json:"read,omitempty"`
	Username        string          `json:"username,omitempty"`
	CreatedAt       time.Time       `json:"created_at,omitempty"`
	UpdatedAt       time.Time       `json:"updated_at,omitempty"`
	Online          bool            `json:"online,omitempty"`
}

type ParentMessages struct {
	ID       uint   `json:"ID,omitempty"`
	Username string `json:"username,omitempty"`
	Text     string `json:"text,omitempty"`
}

type ResponseChat struct {
	ID          uint              `json:"id,omitempty"`
	User1ID     uuid.UUID         `json:"user1_id,omitempty"`
	Name        string            `json:"name,omitempty"`
	User1Name   string            `json:"user1Name,omitempty"`
	User2Name   string            `json:"user2Name,omitempty"`
	LastMessage string            `json:"lastMessage,omitempty"`
	User2ID     uuid.UUID         `json:"user2_id,omitempty"`
	Messages    []ResponseMessage `json:"messages,omitempty"`
	CreatedAt   time.Time         `json:"created_at,omitempty"`
	UpdatedAt   time.Time         `json:"updated_at,omitempty"`
	Class       string            `json:"class,omitempty"`
}

func FilterMessageRecord(message *Message) ResponseMessage {
	var parentMessage *ParentMessages
	if message.ParentMessage != nil {
		parentMessage = &ParentMessages{
			ID:       message.ParentMessage.ID,
			Username: message.ParentMessage.Username,
			Text:     message.ParentMessage.Text,
		}
	}

	var reactions []*ChatReaction
	for _, reaction := range message.Reactions {
		reactionPointer := &ChatReaction{
			ID:        reaction.ID,
			Emoji:     reaction.Emoji,
			UserID:    reaction.UserID,
			MessageID: reaction.MessageID,
			Username:  reaction.Username,
			CreatedAt: reaction.CreatedAt,
			UpdatedAt: reaction.UpdatedAt,
		}
		reactions = append(reactions, reactionPointer)
	}
	return ResponseMessage{
		ID:              message.ID,
		UserID:          *message.UserID,
		User:            message.User,
		ChatID:          message.ChatID,
		ParentMessageID: message.ParentMessageID,
		ParentMessage:   parentMessage,
		Reaction:        reactions,
		Read:            message.Read,
		Text:            message.Text,
		CreatedAt:       message.CreatedAt,
		UpdatedAt:       message.UpdatedAt,
		Online:          false,
	}
}

func FilterChatRecord(chat *Chat) ResponseChat {
	var responseMessages []ResponseMessage
	for _, message := range chat.Messages {
		responseMessages = append(responseMessages, FilterMessageRecord(message))
	}

	return ResponseChat{
		ID:        chat.ID,
		User1ID:   *chat.User1ID,
		User2ID:   *chat.User2ID,
		Messages:  responseMessages,
		CreatedAt: chat.CreatedAt,
		UpdatedAt: chat.UpdatedAt,
	}
}

type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
