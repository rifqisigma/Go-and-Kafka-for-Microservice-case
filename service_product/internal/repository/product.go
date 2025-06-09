package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"service_product/dto"
	"service_product/entity"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ProductRepo interface {
	GetAllProduct() ([]dto.Product, error)
	GetProduct(id uint) (*dto.Product, error)
	CreateProduct(req *dto.CreateProductReq) (*dto.Product, error)
	UpdateProduct(req *dto.UpdateProductReq) (*dto.Product, error)
	DeleteProduct(id uint) error

	//kafka
	GetProductByStoreId(storeId uint) ([]dto.ProductKafka, error)
	ProductValidation(productId uint) (*dto.ValidationProductKafka, error)
	WaitForResponse(correlationID string, out interface{}) error
}

type productRepo struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewStoreRepo(db *gorm.DB, redis *redis.Client) ProductRepo {
	return &productRepo{db, redis}
}

var ctx = context.Background()

func (r *productRepo) CreateProduct(req *dto.CreateProductReq) (*dto.Product, error) {
	newProduct := entity.Product{
		Name:    req.Name,
		StoreID: req.StoreID,
		Stock:   req.Stock,
	}

	if err := r.db.Create(&newProduct).Error; err != nil {
		return nil, err
	}

	key := fmt.Sprintf("product:%d", newProduct.ID)

	_, err := r.redis.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(ctx, key, map[string]interface{}{
			"name":       newProduct.Name,
			"store_id":   newProduct.StoreID,
			"stock":      newProduct.Stock,
			"created_at": newProduct.CreatedAt,
		})
		pipe.Expire(ctx, key, 30*time.Minute)
		pipe.Del(ctx, "products:all")
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("redis: %v", err)
	}

	return &dto.Product{
		ID:        newProduct.ID,
		Name:      newProduct.Name,
		Stock:     newProduct.Stock,
		CreatedAt: newProduct.CreatedAt,
	}, nil
}

func (r *productRepo) UpdateProduct(req *dto.UpdateProductReq) (*dto.Product, error) {
	if err := r.db.Model(&entity.Product{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
		"name":  req.Name,
		"stock": req.Stock,
	}).Error; err != nil {
		return nil, err
	}

	key := fmt.Sprintf("product:%d", req.ID)

	_, err := r.redis.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(ctx, key,
			"name", req.Name,
			"stock", req.Stock,
		)
		pipe.Del(ctx, "products:all")
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("redis: %v", err)
	}

	return &dto.Product{
		ID:      req.ID,
		StoreID: req.StoreID,
		Name:    req.Name,
		Stock:   req.Stock,
	}, nil
}

func (r *productRepo) DeleteProduct(id uint) error {
	tx := r.db.Begin()
	// if err := tx.Model(&model.CartItem{}).Where("product_id = ?", id).Update("is_product_deleted", true).Error; err != nil {
	// 	tx.Rollback()
	// 	return err
	// }
	if err := tx.Delete(&entity.Product{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	key := fmt.Sprintf("product:%d", id)

	_, err := r.redis.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HDel(ctx, key)
		pipe.Del(ctx, "products:all")
		return nil
	})
	if err != nil {
		return fmt.Errorf("redis: %v", err)
	}

	return nil
}

func (r *productRepo) GetProduct(id uint) (*dto.Product, error) {
	key := fmt.Sprintf("product:%d", id)

	var product entity.Product
	data, err := r.redis.HGetAll(ctx, key).Result()
	if err == nil {
		stock, _ := strconv.Atoi(data["stock"])
		storeID, _ := strconv.Atoi(data["store_id"])
		createdAt, _ := time.Parse(time.RFC3339, data["created_at"])

		product := dto.Product{
			ID:        id,
			Name:      data["name"],
			Stock:     stock,
			StoreID:   uint(storeID),
			CreatedAt: createdAt,
		}

		fmt.Println("data dari redis")
		return &product, nil
	}

	if err := r.db.First(&product, id).Error; err != nil {
		return nil, err
	}

	fmt.Println("data dari mysql")
	return &dto.Product{
		ID:        product.ID,
		StoreID:   product.StoreID,
		Name:      product.Name,
		Stock:     product.Stock,
		CreatedAt: product.CreatedAt,
	}, nil
}

func (r *productRepo) GetAllProduct() ([]dto.Product, error) {
	cached, err := r.redis.Get(ctx, "products:all").Result()
	if err == nil {
		var products []dto.Product
		if json.Unmarshal([]byte(cached), &products) == nil {
			fmt.Println("data dari redis")
			return products, nil
		}
	}

	var products []entity.Product
	if err := r.db.Find(&products).Error; err != nil {
		return nil, err
	}

	var result []dto.Product
	for _, p := range products {
		result = append(result, dto.Product{
			ID:        p.ID,
			StoreID:   p.StoreID,
			Name:      p.Name,
			Stock:     p.Stock,
			CreatedAt: p.CreatedAt,
		})
	}

	jsonData, _ := json.Marshal(result)
	if err := r.redis.Set(ctx, "products:all", jsonData, 30*time.Minute); err != nil {
		return nil, fmt.Errorf("redis: %v", err)
	}

	fmt.Println("data dari mysql")
	return result, nil
}

func (r *productRepo) GetProductByStoreId(storeId uint) ([]dto.ProductKafka, error) {
	var products []dto.ProductKafka
	if err := r.db.Model(&entity.Product{}).Select("id, name, stock, created_at").Where("store_id = ?", storeId).Order("created_at ASC").Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}

func (r *productRepo) ProductValidation(productId uint) (*dto.ValidationProductKafka, error) {
	var product entity.Product
	err := r.db.Model(&entity.Product{}).Select("stock").Where("id = ?", productId).First(&product).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &dto.ValidationProductKafka{
			Deleted: true,
			Stock:   0,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return &dto.ValidationProductKafka{
		Deleted: false,
		Stock:   product.Stock,
	}, nil
}

func (u *productRepo) WaitForResponse(correlationID string, out interface{}) error {
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
