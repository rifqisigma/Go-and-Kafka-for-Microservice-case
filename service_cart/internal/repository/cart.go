package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"service_cart/dto"
	"service_cart/entity"
	"service_cart/helper/utils"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type CartRepo interface {

	// cart item
	GetMyCartItems(userId uint) ([]dto.CartItem, error)
	CreateCartItem(req *dto.CreateCartItemReq) error
	UpdateAmountCartItem(req *dto.UpdateAmountCartItemReq) error
	UpdatePaidCartItem(req *dto.UpdatePaidCartItemReq) error
	DeleteCartItem(userId, id uint) error

	//kafka
	UpdateIsDeleteProduct(id uint) error
	WaitForResponse(correlationID string, out interface{}) error
}

type cartRepo struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewCartRepo(db *gorm.DB, redis *redis.Client) CartRepo {
	return &cartRepo{db, redis}
}

var ctx = context.Background()

func (r *cartRepo) CreateCartItem(req *dto.CreateCartItemReq) error {
	newCartItem := entity.CartItem{
		ProductID:      &req.ProductID,
		UserID:         req.UserID,
		PurchaseAmount: req.PurchaseAmount}

	if err := r.db.Model(&entity.CartItem{}).Create(&newCartItem).Error; err != nil {
		return err
	}

	key := fmt.Sprintf("user:%d:cart_items", req.UserID)
	deleted, err := r.redis.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	if deleted == 0 {
		fmt.Println("Key tidak ditemukan di Redis")
	}
	return nil
}

func (r *cartRepo) UpdateAmountCartItem(req *dto.UpdateAmountCartItemReq) error {
	if err := r.db.Model(&entity.CartItem{}).Where("id = ?", req.ID).Update("purchase_amount", req.PurchaseAmount).Error; err != nil {
		return err
	}

	key := fmt.Sprintf("user:%d:cart_items", req.UserID)
	deleted, err := r.redis.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	if deleted == 0 {
		fmt.Println("Key tidak ditemukan di Redis")
	}
	return nil
}

func (r *cartRepo) UpdatePaidCartItem(req *dto.UpdatePaidCartItemReq) error {
	if err := r.db.Model(&entity.CartItem{}).Where("user_id = ? AND id = ? AND is_product_deleted = ?", req.UserID, req.ID, false).Update("is_paid", true).Error; err != nil {
		return err
	}

	key := fmt.Sprintf("user:%d:cart_items", req.UserID)
	deleted, err := r.redis.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	if deleted == 0 {
		fmt.Println("Key tidak ditemukan di Redis")
	}
	return nil
}

func (r *cartRepo) DeleteCartItem(userId, id uint) error {
	if err := r.db.Model(&entity.CartItem{}).Where("id = ?", id).Delete(&entity.CartItem{}).Error; err != nil {
		return err
	}

	key := fmt.Sprintf("user:%d:cart_items", userId)
	deleted, err := r.redis.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	if deleted == 0 {
		fmt.Println("Key tidak ditemukan di Redis")
	}
	return nil
}

func (r *cartRepo) GetMyCartItems(userId uint) ([]dto.CartItem, error) {
	key := fmt.Sprintf("user:%d:cart_items", userId)
	cachedData, err := r.redis.Get(ctx, key).Result()
	if err == nil && cachedData != "" {
		var cachedItems []dto.CartItem
		if err := json.Unmarshal([]byte(cachedData), &cachedItems); err == nil {
			return cachedItems, nil
		}
	}

	var items []dto.CartItem
	if err := r.db.Where("user_id = ?", userId).Find(&items).Error; err != nil {

		return nil, err
	}

	jsonData, _ := json.Marshal(items)

	if err := r.redis.Set(ctx, key, jsonData, 30*time.Minute).Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *cartRepo) UpdateIsDeleteProduct(id uint) error {
	tx := r.db.Begin()
	var userId []uint
	err := tx.Model(&entity.CartItem{}).Where("product_id = ? AND is_product_deleted = ?", id, false).Pluck("user_id", &userId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return utils.ErrProductDeleted
	}
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&entity.CartItem{}).Update("is_product_deleted", true).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, u := range userId {
		key := fmt.Sprintf("user:%d:cart_items", u)
		deleted, err := r.redis.Del(ctx, key).Result()
		if err != nil {
			return err
		}
		if deleted == 0 {
			fmt.Println("Key tidak ditemukan di Redis")
		}
	}

	return nil
}

func (u *cartRepo) WaitForResponse(correlationID string, out interface{}) error {
	key := fmt.Sprintf("response:%s", correlationID)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for Kafka response")
		case <-ticker.C:
			val, err := u.redis.Get(ctx, key).Result()
			if err == redis.Nil {
				continue
			}
			if err != nil {
				return err
			}
			return json.Unmarshal([]byte(val), out)
		}
	}
}
