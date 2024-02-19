package models

import (
	"github.com/go-playground/validator/v10"
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
}

type GroupMessage struct {
	ID              uint       `gorm:"primaryKey"`
	GroupID         uint       `gorm:"not null"`
	Group           *Group     `gorm:"foreignKey:GroupID"`
	User            *User      `gorm:"foreignKey:UserID"`
	UserID          *uuid.UUID `gorm:"type:uuid"`
	ParentMessage   *Message   `gorm:"foreignKey:ParentMessageID"`
	ParentMessageID *uint
	Read            bool      `gorm:"default:false"`
	Text            string    `gorm:"not null"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
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
	ID              uint         `json:"id,omitempty"`
	GroupID         uint         `json:"group_id,omitempty"`
	User            UserResponse `json:"user,omitempty"`
	ParentMessageID *uint        `json:"parent_message_id,omitempty"`
	Read            bool         `json:"read"`
	Text            string       `json:"text,omitempty"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

func FilterGroupMessageRecord(GroupMessage *GroupMessage) GroupMessageResponse {
	return GroupMessageResponse{
		ID:              GroupMessage.ID,
		GroupID:         GroupMessage.GroupID,
		User:            FilterUserRecord(GroupMessage.User),
		ParentMessageID: GroupMessage.ParentMessageID,
		Read:            GroupMessage.Read,
		Text:            GroupMessage.Text,
		CreatedAt:       GroupMessage.CreatedAt,
		UpdatedAt:       GroupMessage.UpdatedAt,
	}
}

type ErrorResponses struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value,omitempty"`
}

func ValidateStructs[T any](payload T) []*ErrorResponse {
	var errors []*ErrorResponse
	err := validate.Struct(payload)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element ErrorResponse
			element.Field = err.StructNamespace()
			element.Tag = err.Tag()
			element.Value = err.Param()
			errors = append(errors, &element)
		}
	}
	return errors
}
