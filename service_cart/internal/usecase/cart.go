package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"service_cart/dto"
	"service_cart/helper/utils"
	"service_cart/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/sony/gobreaker"
)

type CartUsecase interface {
	GetMyCartItems(userId uint) ([]dto.CartItem, error)
	CreateCartItem(req *dto.CreateCartItemReq) error
	UpdateAmountCartItem(req *dto.UpdateAmountCartItemReq) error
	UpdatePaidCartItem(req *dto.UpdatePaidCartItemReq) error
	DeleteCartItem(userId, id uint) error
	UpdateIsDeleteProduct(id uint) error

	//kafka
	WriteKafkaMessage(topic string, key string, payload interface{}) error
}

type cartUsecase struct {
	cartRepo     repository.CartRepo
	kafka        map[string]*kafka.Writer
	writeBreaker *gobreaker.CircuitBreaker
}

func NewCartUsecase(cartRepo repository.CartRepo, kafka map[string]*kafka.Writer) CartUsecase {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "ProducerBreaker",
		MaxRequests: 5,
		Interval:    30 * time.Second,
		Timeout:     5 * time.Second,
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Printf("[Circuit Breaker: %s] status berubah dari %s ➜ %s\n", name, from.String(), to.String())
		},
	})
	return &cartUsecase{cartRepo, kafka, cb}
}

func (u *cartUsecase) WriteKafkaMessage(topic string, key string, payload interface{}) error {
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

func (u *cartUsecase) GetMyCartItems(userId uint) ([]dto.CartItem, error) {
	return u.cartRepo.GetMyCartItems(userId)
}

func (u *cartUsecase) CreateCartItem(req *dto.CreateCartItemReq) error {
	corrId := uuid.NewString()

	payload := map[string]interface{}{
		"correlation_id": corrId,
		"product_id":     req.ProductID,
	}
	if err := u.WriteKafkaMessage("validation-product-request", corrId, payload); err != nil {
		return utils.ErrFailedKafkaWrite
	}

	var validation dto.ValidationProductKafka
	if err := u.cartRepo.WaitForResponse(corrId, &validation); err != nil {
		return err
	}

	if !validation.Deleted {
		return utils.ErrProductDeleted
	}

	if validation.Stock < req.PurchaseAmount {
		return utils.ErrStocknotEnough
	}
	return u.cartRepo.CreateCartItem(req)
}

func (u *cartUsecase) UpdateAmountCartItem(req *dto.UpdateAmountCartItemReq) error {
	corrId := uuid.NewString()

	payload := map[string]interface{}{
		"correlation_id": corrId,
		"product_id":     req.ProductID,
	}
	if err := u.WriteKafkaMessage("validation-product-request", corrId, payload); err != nil {
		return utils.ErrFailedKafkaWrite
	}

	var validation dto.ValidationProductKafka
	if err := u.cartRepo.WaitForResponse(corrId, &validation); err != nil {
		return err
	}

	if !validation.Deleted {
		return utils.ErrProductDeleted
	}

	if validation.Stock < req.PurchaseAmount {
		return utils.ErrStocknotEnough
	}
	return u.cartRepo.UpdateAmountCartItem(req)
}

func (u *cartUsecase) UpdatePaidCartItem(req *dto.UpdatePaidCartItemReq) error {
	corrId := uuid.NewString()

	payload := map[string]interface{}{
		"correlation_id": corrId,
		"product_id":     req.ProductID,
	}
	if err := u.WriteKafkaMessage("validation-product-request", corrId, payload); err != nil {
		return utils.ErrFailedKafkaWrite
	}

	var validation dto.ValidationProductKafka
	if err := u.cartRepo.WaitForResponse(corrId, &validation); err != nil {
		return err
	}

	if !validation.Deleted {
		return utils.ErrProductDeleted
	}

	if validation.Stock < req.PurchaseAmount {
		return utils.ErrStocknotEnough
	}

	message, _ := json.Marshal(&req)
	payloadtwo := map[string]interface{}{
		"correlation_id": corrId,
		"email":          req.Email,
		"service":        "cart",
		"action":         "paid",
		"message":        message,
	}

	if err := u.WriteKafkaMessage("notification-request", corrId, payloadtwo); err != nil {
		return utils.ErrFailedKafkaWrite
	}
	return u.cartRepo.UpdatePaidCartItem(req)
}

func (u *cartUsecase) DeleteCartItem(userId, id uint) error {
	return u.cartRepo.DeleteCartItem(userId, id)
}

func (u *cartUsecase) UpdateIsDeleteProduct(id uint) error {
	return u.cartRepo.UpdateIsDeleteProduct(id)
}
