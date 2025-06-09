package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"service_store/dto"
	"service_store/entity"
	"service_store/helper/utils"
	"time"

	"github.com/redis/go-redis/v9"

	"gorm.io/gorm"
)

type StoreRepo interface {
	IsUserAdminStore(userId, storeId uint) (bool, error)
	GetMyStore(id uint) (*dto.Store, error)
	GetAllStore() ([]dto.Store, error)
	CreateStore(req *dto.CreateStoreReq) (*dto.Store, error)
	UpdateStore(req *dto.UpdateStoreReq) (*dto.Store, error)
	DeleteStore(id uint) error

	//kafka
	WaitForResponse(correlationID string, out interface{}) error
}

type storeRepo struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewStoreRepo(db *gorm.DB, redis *redis.Client) StoreRepo {
	return &storeRepo{db, redis}
}

var ctx = context.Background()

func (r *storeRepo) IsUserAdminStore(userId, storeId uint) (bool, error) {
	var count int64
	if err := r.db.Model(&entity.Store{}).Where("id = ? AND admin_id = ?", storeId, userId).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *storeRepo) GetMyStore(id uint) (*dto.Store, error) {
	// Ambil data store
	var store entity.Store
	if err := r.db.Model(&entity.Store{}).Where("id = ?", id).First(&store).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNoStore
		}
		return nil, err
	}

	return &dto.Store{
		ID:        store.ID,
		Name:      store.Name,
		AdminID:   store.AdminID,
		CreatedAt: store.CreatedAt,
	}, nil
}

func (r *storeRepo) GetAllStore() ([]dto.Store, error) {
	key := fmt.Sprintln("store:all")

	cachedData, err := r.redis.Get(ctx, key).Result()
	if err == nil && cachedData != "" {
		var cachedShops []dto.Store
		if err := json.Unmarshal([]byte(cachedData), &cachedShops); err == nil {
			log.Println("data dari redis")
			return cachedShops, nil
		}
	}

	log.Println("data dari mysql")
	var shops []dto.Store
	if err := r.db.Model(&entity.Store{}).Select("id", "name", "admin_id", "created_at").Find(&shops).Error; err != nil {

		return nil, err
	}

	jsonData, _ := json.Marshal(shops)
	if err := r.redis.Set(ctx, key, jsonData, 15*time.Minute).Err(); err != nil {
		return nil, fmt.Errorf("redis: %v", err)
	}

	return shops, nil
}

func (r *storeRepo) CreateStore(req *dto.CreateStoreReq) (*dto.Store, error) {
	newStore := entity.Store{
		Name:    req.Name,
		AdminID: req.AdminID,
	}

	if err := r.db.Model(&entity.Store{}).Create(&newStore).Error; err != nil {
		return nil, err
	}

	key := fmt.Sprintln("store:all")
	if err := r.redis.Del(ctx, key).Err(); err != nil {
		return nil, err
	}

	return &dto.Store{
		ID:        newStore.ID,
		AdminID:   newStore.AdminID,
		Name:      newStore.Name,
		CreatedAt: time.Now(),
	}, nil
}

func (r *storeRepo) UpdateStore(req *dto.UpdateStoreReq) (*dto.Store, error) {
	if err := r.db.Model(&entity.Store{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
		"name": req.Name,
	}).Error; err != nil {
		return nil, err
	}

	key := fmt.Sprintln("store:all")
	if err := r.redis.Del(ctx, key).Err(); err != nil {
		return nil, err
	}
	return &dto.Store{
		ID:      req.ID,
		AdminID: req.UserID,
		Name:    req.Name,
	}, nil
}

func (r *storeRepo) DeleteStore(id uint) error {
	if err := r.db.Model(&entity.Store{}).Where("id = ?", id).Delete(&entity.Store{}).Error; err != nil {
		return err
	}

	key := fmt.Sprintln("store:all")
	if err := r.redis.Del(ctx, key).Err(); err != nil {
		return err
	}
	return nil
}

func (u *storeRepo) WaitForResponse(correlationID string, out interface{}) error {
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
