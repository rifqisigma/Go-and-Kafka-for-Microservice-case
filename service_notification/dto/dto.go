package dto

import "time"

type SendEmail struct {
	ToEmail  string
	Header   string
	ActionId string
	Desc     string
}

type Store struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	AdminID   uint      `json:"admin_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Product struct {
	StoreID   uint      `json:"store_id"`
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Stock     int       `json:"stock"`
	CreatedAt time.Time `json:"created_at"`
}

type UpdatePaidCartItemReq struct {
	ID             uint `json:"cart_id"`
	ProductID      uint `json:"product_id"`
	PurchaseAmount int  `json:"purchase_amount"`
}
