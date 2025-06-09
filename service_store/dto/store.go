package dto

import "time"

//store
type CreateStoreReq struct {
	Email   string `json:"-"`
	AdminID uint   `json:"-"`
	Name    string `json:"name"`
}

type UpdateStoreReq struct {
	Email  string `json:"-"`
	UserID uint   `json:"-"`
	ID     uint   `json:"-"`
	Name   string `json:"name"`
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

type StoreAndProduct struct {
	Store   Store     `json:"store"`
	Product []Product `json:"products"`
}
