package entity

import "time"

type CartItem struct {
	ID               uint `gorm:"primaryKey"`
	PurchaseAmount   int  `gorm:"not null"`
	IsPaid           bool `gorm:"default:false"`
	CreatedAt        time.Time
	IsProductDeleted bool  `gorm:"default:false"`
	UserID           uint  `gorm:"index"`
	ProductID        *uint `gorm:"index"`
}
