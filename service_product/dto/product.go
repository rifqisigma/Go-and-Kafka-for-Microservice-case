package dto

import "time"

type CreateProductReq struct {
	Email   string `json:"-"`
	UserID  uint   `json:"-"`
	StoreID uint   `json:"-"`
	Name    string `json:"name"`
	Stock   int    `json:"stock"`
}

type UpdateProductReq struct {
	Email   string `json:"-"`
	ID      uint   `json:"-"`
	UserID  uint   `json:"-"`
	StoreID uint   `json:"-"`
	Name    string `json:"name"`
	Stock   int    `json:"stock"`
}

type Product struct {
	StoreID   uint      `json:"store_id"`
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Stock     int       `json:"stock"`
	CreatedAt time.Time `json:"created_at"`
}

type ProductKafka struct {
	Name      string
	Stock     int
	StoreID   uint
	CreatedAt time.Time
}

type ValidationProductKafka struct {
	ProductId uint
	Stock     int
	Deleted   bool
}
