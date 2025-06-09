package entity

import "time"

type Store struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique;not null"`

	//admin
	AdminID uint `gorm:"index;unique"`

	//product
	CreatedAt time.Time
}
