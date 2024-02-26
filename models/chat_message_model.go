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
	ID              uint       `gorm:"primaryKey"`
	User            *User      `gorm:"foreignKey:UserID"`
	UserID          *uuid.UUID `gorm:"type:uuid"`
	Chat            Chat       `gorm:"foreignKey:ChatID"`
	ChatID          uint
	ParentMessage   *Message `gorm:"foreignKey:ParentMessageID"`
	ParentMessageID *uint
	StickerID       *uint
	Sticker         Sticker         `gorm:"foreignKey:StickerID"`
	Read            bool            `gorm:"default:false"`
	Text            string          `gorm:"not null"`
	Reactions       []*ChatReaction `gorm:"many2many:message_reactions;"`
	CreatedAt       time.Time       `gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime"`
}
type ChatReaction struct {
	ID        uint       `gorm:"primaryKey"`
	Emoji     string     `gorm:"not null"`
	UserID    *uuid.UUID `gorm:"type:uuid;not null"`
	MessageID uint       `gorm:"not null"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
}

type ResponseMessage struct {
	ID              uint         `json:"id,omitempty"`
	UserID          uuid.UUID    `json:"user_id,omitempty"`
	User            UserResponse `json:"user,omitempty"`
	ChatID          uint         `json:"chat_id,omitempty"`
	ParentMessageID *uint        `json:"parent_message_id,omitempty"`
	ParentMessage   *Message     `json:"parent_message,omitempty"`
	Text            string       `json:"text,omitempty"`
	Username        string       `json:"username,omitempty"`
	CreatedAt       time.Time    `json:"created_at,omitempty"`
	UpdatedAt       time.Time    `json:"updated_at,omitempty"`
}

type ResponseChat struct {
	ID        uint              `json:"id,omitempty"`
	User1ID   uuid.UUID         `json:"user1_id,omitempty"`
	User2ID   uuid.UUID         `json:"user2_id,omitempty"`
	Messages  []ResponseMessage `json:"messages,omitempty"`
	CreatedAt time.Time         `json:"created_at,omitempty"`
	UpdatedAt time.Time         `json:"updated_at,omitempty"`
}

func FilterMessageRecord(message *Message) ResponseMessage {

	return ResponseMessage{
		ID:              message.ID,
		UserID:          *message.UserID,
		ChatID:          message.ChatID,
		ParentMessageID: message.ParentMessageID,
		Text:            message.Text,
		CreatedAt:       message.CreatedAt,
		UpdatedAt:       message.UpdatedAt,
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
