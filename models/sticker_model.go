package models

type StickerPack struct {
	ID       uint       `gorm:"primaryKey"`
	Name     string     `gorm:"unique;not null"`
	Stickers []*Sticker `gorm:"foreignKey:PackID"`
}

type Sticker struct {
	ID     uint         `gorm:"primaryKey"`
	PackID uint         `gorm:"not null"`
	Pack   *StickerPack `gorm:"not null"`
	URL    string       `gorm:"not null"`
}
