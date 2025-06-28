package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"service_store/dto"
	"service_store/helper/utils"
	"time"

	"service_store/internal/repository"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/sony/gobreaker"
)

type StoreUsecase interface {

	//store
	GetAllStore() ([]dto.Store, error)
	CreateStore(req *dto.CreateStoreReq) error
	UpdateStore(req *dto.UpdateStoreReq) error
	DeleteStore(storeId, userId uint, email string) error

	//kafka
	GetMyStore(storeId uint) (*dto.StoreAndProduct, error)
	SendValidationResponse(userId, storeId uint, correlationID string) error
	WriteKafkaMessage(topic string, key string, payload interface{}) error
}

type storeUsecase struct {
	storeRepo    repository.StoreRepo
	kafka        map[string]*kafka.Writer
	writeBreaker *gobreaker.CircuitBreaker
}

func NewStoreUsecase(storeRepo repository.StoreRepo, kafka map[string]*kafka.Writer) StoreUsecase {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "ProducerBreaker",
		MaxRequests: 5,
		Interval:    30 * time.Second,
		Timeout:     5 * time.Second,
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Printf("[Circuit Breaker: %s] status berubah dari %s ➜ %s\n", name, from.String(), to.String())
		},
	})

	return &storeUsecase{storeRepo, kafka, cb}
}

func (u *storeUsecase) WriteKafkaMessage(topic string, key string, payload interface{}) error {
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

func (u *storeUsecase) GetAllStore() ([]dto.Store, error) {
	return u.storeRepo.GetAllStore()
}

func (u *storeUsecase) GetMyStore(storeId uint) (*dto.StoreAndProduct, error) {
	store, err := u.storeRepo.GetMyStore(storeId)
	if err != nil {
		return nil, err
	}
	corrID := uuid.NewString()

	payload := map[string]interface{}{
		"store_id":       storeId,
		"correlation_id": corrID,
	}

	if err := u.WriteKafkaMessage("products-request", corrID, payload); err != nil {
		return nil, utils.ErrFailedKafkaWrite
	}

	var productData []dto.Product
	if err := u.storeRepo.WaitForResponse(corrID, &productData); err != nil {
		return nil, err
	}

	return &dto.StoreAndProduct{
		Store:   *store,
		Product: productData,
	}, nil
}

func (u *storeUsecase) CreateStore(req *dto.CreateStoreReq) error {
	corrId := uuid.NewString()

	store, err := u.storeRepo.CreateStore(req)
	if err != nil {
		return err
	}

	message, _ := json.Marshal(&store)

	payload := map[string]interface{}{
		"correlation_id": corrId,
		"email":          req.Email,
		"service":        "store",
		"action":         "create",
		"message":        message,
	}

	if err := u.WriteKafkaMessage("notification-request", corrId, payload); err != nil {
		return err
	}

	return nil
}

func (u *storeUsecase) UpdateStore(req *dto.UpdateStoreReq) error {
	corrId := uuid.NewString()

	valid, err := u.storeRepo.IsUserAdminStore(req.UserID, req.ID)
	if err != nil {
		return err
	}
	if !valid {
		return utils.ErrNotAdmin
	}

	store, err := u.storeRepo.UpdateStore(req)
	if err != nil {
		return err
	}
	message, _ := json.Marshal(&store)

	payload := map[string]interface{}{
		"correlation_id": corrId,
		"email":          req.Email,
		"service":        "store",
		"action":         "update",
		"message":        message,
	}

	if err := u.WriteKafkaMessage("notification-request", corrId, payload); err != nil {
		return utils.ErrFailedKafkaWrite
	}

	return nil
}

func (u *storeUsecase) DeleteStore(storeId, userId uint, email string) error {
	corrId := uuid.NewString()

	valid, err := u.storeRepo.IsUserAdminStore(userId, storeId)
	if err != nil {
		return err
	}
	if !valid {
		return utils.ErrNotAdmin
	}

	if err := u.storeRepo.DeleteStore(storeId); err != nil {
		return err
	}

	if err := u.WriteKafkaMessage("notification-request", corrId, storeId); err != nil {
		return utils.ErrFailedKafkaWrite
	}

	return nil

}

func (u *storeUsecase) SendValidationResponse(userId, storeId uint, correlationID string) error {
	valid, err := u.storeRepo.IsUserAdminStore(userId, storeId)
	if err != nil {
		return err
	}
	payload := map[string]interface{}{
		"correlation_id": correlationID,
		"is_valid":       valid,
	}

	if err := u.WriteKafkaMessage("store-validation-response", correlationID, payload); err != nil {
		return err
	}

	return nil
}
