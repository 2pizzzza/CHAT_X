package models

import (
	"github.com/google/uuid"
	"time"
)

type Group struct {
	ID           uint            `gorm:"primaryKey"`
	Name         string          `gorm:"unique;not null"`
	Description  string          `gorm:"not null"`
	PhotoURL     string          `gorm:"not null"`
	Admins       []*User         `gorm:"many2many:group_admins;"`
	Participants []*User         `gorm:"many2many:group_participants;"`
	CreatedAt    time.Time       `gorm:"autoCreateTime"`
	UpdatedAt    time.Time       `gorm:"autoUpdateTime"`
	Message      []*GroupMessage `gorm:"foreignKey:GroupID"`
	Target       string          `gorm:"default:personal"`
}

type GroupMessage struct {
	ID              uint                 `gorm:"primaryKey"`
	GroupID         uint                 `gorm:"not null"`
	Group           *Group               `gorm:"foreignKey:GroupID"`
	User            *User                `gorm:"foreignKey:UserID"`
	UserID          *uuid.UUID           `gorm:"type:uuid"`
	ParentMessage   *ParentMessagesGroup `gorm:"-"`
	ParentMessageID *uint
	StickerID       *uint
	Sticker         Sticker     `gorm:"foreignKey:StickerID"`
	Read            bool        `gorm:"default:false"`
	Text            string      `gorm:"not null"`
	Reactions       []*Reaction `gorm:"many2many:group_message_reactions;"`
	CreatedAt       time.Time   `gorm:"autoCreateTime"`
	UpdatedAt       time.Time   `gorm:"autoUpdateTime"`
}

type Reaction struct {
	ID        uint       `gorm:"primaryKey"`
	Emoji     string     `gorm:"not null"`
	UserID    *uuid.UUID `gorm:"type:uuid;not null"`
	MessageID uint       `gorm:"not null"`
	Username  string     `json:"username,omitempty"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
}

type GroupResponse struct {
	ID           uint      `json:"id,omitempty"`
	Name         string    `json:"name,omitempty"`
	Description  string    `json:"description,omitempty"`
	PhotoURL     string    `json:"photo_url,omitempty"`
	Admins       []*User   `json:"admins,omitempty"`
	Participants []*User   `json:"participants,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func FilterGroupRecord(group *Group) GroupResponse {
	return GroupResponse{
		ID:           group.ID,
		Name:         group.Name,
		Description:  group.Description,
		PhotoURL:     group.PhotoURL,
		Admins:       group.Admins,
		Participants: group.Participants,
		CreatedAt:    group.CreatedAt,
		UpdatedAt:    group.UpdatedAt,
	}
}

type GroupMessageResponse struct {
	ID              uint                 `json:"id,omitempty"`
	GroupID         uint                 `json:"group_id,omitempty"`
	UserID          uuid.UUID            `json:"user_id,omitempty"`
	ParentMessageID *uint                `json:"parent_message_id,omitempty"`
	ParentMessage   *ParentMessagesGroup `json:"parent_message,omitempty"`
	Read            bool                 `json:"read"`
	Reaction        []*Reaction          `json:"reaction,omitempty"`
	Username        string               `json:"username,omitempty"`
	Text            string               `json:"text,omitempty"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

type ParentMessagesGroup struct {
	ID       uint   `json:"ID,omitempty"`
	Username string `json:"username,omitempty"`
	Text     string `json:"text,omitempty"`
}

func FilterGroupMessageRecord(groupMessage *GroupMessage) GroupMessageResponse {
	var parentMessage *ParentMessagesGroup
	if groupMessage.ParentMessage != nil {
		parentMessage = &ParentMessagesGroup{
			ID:       groupMessage.ParentMessage.ID,
			Username: groupMessage.ParentMessage.Username,
			Text:     groupMessage.ParentMessage.Text,
		}
	}

	var reactions []*Reaction
	for _, reaction := range groupMessage.Reactions {
		reactionPointer := &Reaction{
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

	return GroupMessageResponse{
		ID:              groupMessage.ID,
		UserID:          *groupMessage.UserID,
		GroupID:         groupMessage.GroupID,
		ParentMessageID: groupMessage.ParentMessageID,
		ParentMessage:   parentMessage,
		Reaction:        reactions,
		Read:            groupMessage.Read,
		Text:            groupMessage.Text,
		CreatedAt:       groupMessage.CreatedAt,
		UpdatedAt:       groupMessage.UpdatedAt,
	}
}

type ErrorResponses struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value,omitempty"`
}

//func ValidateStructs[T any](payload T) []*ErrorResponse {
//	var errors []*ErrorResponse
//	err := validate.Struct(payload)
//	if err != nil {
//		for _, err := range err.(validator.ValidationErrors) {
//			var element ErrorResponse
//			element.Field = err.StructNamespace()
//			element.Tag = err.Tag()
//			element.Value = err.Param()
//			errors = append(errors, &element)
//		}
//	}
//	return errors
//}
