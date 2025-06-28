package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"service_product/dto"
	"service_product/helper/utils"
	"service_product/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/sony/gobreaker"
)

type ProductUsecase interface {

	//product
	GetAllProduct() ([]dto.Product, error)
	GetProduct(id uint) (*dto.Product, error)
	CreateProduct(req *dto.CreateProductReq) error
	UpdateProduct(req *dto.UpdateProductReq) error
	DeleteProduct(userId, storeId, id uint, email string) error

	//kafka
	SendProductsResponse(storeId uint, correlation_id string) error
	SendValidationCartResponse(productId uint, correlation_id string) error
	WriteKafkaMessage(topic string, key string, payload interface{}) error
}

type productUsecase struct {
	productRepo  repository.ProductRepo
	kafka        map[string]*kafka.Writer
	writeBreaker *gobreaker.CircuitBreaker
}

func NewProductUsecase(productRepo repository.ProductRepo, kafka map[string]*kafka.Writer) ProductUsecase {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "ProducerBreaker",
		MaxRequests: 5,
		Interval:    30 * time.Second,
		Timeout:     5 * time.Second,
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Printf("[Circuit Breaker: %s] status berubah dari %s ➜ %s\n", name, from.String(), to.String())
		},
	})
	return &productUsecase{productRepo, kafka, cb}
}

func (u *productUsecase) WriteKafkaMessage(topic string, key string, payload interface{}) error {
	writer, ok := u.kafka[topic]
	if !ok {
		return utils.ErrNoTopic
	}

	value, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(key),
		Value: value,
	}

	_, err = u.writeBreaker.Execute(func() (interface{}, error) {
		return nil, writer.WriteMessages(context.Background(), msg)
	})

	if err != nil {
		return fmt.Errorf("kafka write failed or circuit open: %w", err)
	}

	return nil
}

func (u *productUsecase) CreateProduct(req *dto.CreateProductReq) error {
	corrID := uuid.NewString()

	payload := map[string]interface{}{
		"store_id":       req.StoreID,
		"user_id":        req.UserID,
		"correlation_id": corrID,
	}
	if err := u.WriteKafkaMessage("store-validation-request", corrID, payload); err != nil {
		return err
	}

	var isValid bool
	if err := u.productRepo.WaitForResponse(corrID, &isValid); err != nil {
		return err
	}
	if !isValid {
		return utils.ErrNotAdmin
	}

	product, err := u.productRepo.CreateProduct(req)
	if err != nil {
		return err
	}

	message, _ := json.Marshal(&product)
	payloadtwo := map[string]interface{}{
		"correlation_id": corrID,
		"email":          req.Email,
		"service":        "product",
		"action":         "create",
		"message":        message,
	}
	if err := u.WriteKafkaMessage("notification-request", corrID, payloadtwo); err != nil {
		return err
	}

	return nil
}

func (u *productUsecase) UpdateProduct(req *dto.UpdateProductReq) error {
	corrID := uuid.NewString()

	payload := map[string]interface{}{
		"store_id":       req.StoreID,
		"user_id":        req.UserID,
		"correlation_id": corrID,
	}
	if err := u.WriteKafkaMessage("store-validation-request", corrID, payload); err != nil {
		return err
	}

	var isValid bool
	if err := u.productRepo.WaitForResponse(corrID, &isValid); err != nil {
		return err
	}
	if !isValid {
		return utils.ErrNotAdmin
	}
	product, err := u.productRepo.UpdateProduct(req)
	if err != nil {
		return err
	}

	message, _ := json.Marshal(&product)
	payloadtwo := map[string]interface{}{
		"correlation_id": corrID,
		"email":          req.Email,
		"service":        "product",
		"action":         "update",
		"message":        message,
	}
	if err := u.WriteKafkaMessage("notification-request", corrID, payloadtwo); err != nil {
		return err
	}
	return nil
}

func (u *productUsecase) DeleteProduct(userId, storeId, id uint, email string) error {
	corrID := uuid.NewString()

	payload := map[string]interface{}{
		"store_id":       storeId,
		"user_id":        userId,
		"correlation_id": corrID,
	}
	if err := u.WriteKafkaMessage("store-validation-request", corrID, payload); err != nil {
		return err
	}

	var isValid bool
	if err := u.productRepo.WaitForResponse(corrID, &isValid); err != nil {
		return err
	}
	if !isValid {
		return utils.ErrNotAdmin
	}

	payloadtwo := map[string]interface{}{
		"correlation_id": corrID,
		"deleted":        true,
		"stock":          0,
		"product_id":     id,
	}
	if err := u.WriteKafkaMessage("validation-product-response", corrID, payloadtwo); err != nil {
		return err
	}

	payloadthree := map[string]interface{}{
		"correlation_id": corrID,
		"email":          email,
		"service":        "product",
		"action":         "update",
		"message":        id,
	}
	if err := u.WriteKafkaMessage("notification-request", corrID, payloadthree); err != nil {
		return err
	}
	return u.productRepo.DeleteProduct(id)
}

func (u *productUsecase) GetAllProduct() ([]dto.Product, error) {
	return u.productRepo.GetAllProduct()
}

func (u *productUsecase) GetProduct(id uint) (*dto.Product, error) {
	return u.productRepo.GetProduct(id)
}

func (u *productUsecase) SendProductsResponse(storeId uint, correlation_id string) error {
	products, err := u.productRepo.GetProductByStoreId(storeId)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"correlation_id": correlation_id,
		"data":           products,
	}

	if err := u.WriteKafkaMessage("products-response", correlation_id, payload); err != nil {
		return err
	}

	return nil
}

func (u *productUsecase) SendValidationCartResponse(productId uint, correlation_id string) error {
	result, err := u.productRepo.ProductValidation(productId)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"correlation_id": correlation_id,
		"deleted":        result.Deleted,
		"stock":          result.Stock,
		"product_id":     result.ProductId,
	}

	if err := u.WriteKafkaMessage("validation-product-response", correlation_id, payload); err != nil {
		return err
	}

	return nil

}
