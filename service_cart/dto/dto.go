package dto

import "time"

type CartItem struct {
	ID               uint `json:"id"`
	UserID           uint `json:"user_id"`
	ProductID        uint `json:"product_id"`
	PurchaseAmount   int  `json:"purchase_amount"`
	IsPaid           bool `json:"id_paid"`
	CreatedAt        time.Time
	IsProductDeleted bool `json:"is_product_deleted"`
}

type CreateCartItemReq struct {
	UserID         uint `json:"-"`
	ProductID      uint `json:"-"`
	PurchaseAmount int  `json:"purchase_amount"`
}

type UpdateAmountCartItemReq struct {
	UserID         uint `json:"-"`
	ID             uint `json:"-"`
	ProductID      uint `json:"-"`
	PurchaseAmount int  `json:"purchase_amount"`
}

type UpdatePaidCartItemReq struct {
	Email          string `json:"-"`
	UserID         uint   `json:"-"`
	ID             uint   `json:"-"`
	ProductID      uint   `json:"-"`
	PurchaseAmount int    `json:"purchase_amount"`
}

type ValidationProductKafka struct {
	Stock   int  `json:"stock"`
	Deleted bool `json:"deleted"`
}
