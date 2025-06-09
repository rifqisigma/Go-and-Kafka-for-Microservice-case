package entity

import "time"

type Product struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"not null"`
	Stock int    `gorm:"not null"`

	//store
	StoreID   uint `gorm:"index"`
	CreatedAt time.Time
}
