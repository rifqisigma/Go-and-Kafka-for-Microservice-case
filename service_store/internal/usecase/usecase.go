package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"service_store/dto"
	"service_store/helper/utils"

	"service_store/internal/repository"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
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
	SendNotifRequest(correlationID, email string, message interface{}, service, action string) error
}

type storeUsecase struct {
	storeRepo repository.StoreRepo
	kafka     map[string]*kafka.Writer
}

func NewStoreUsecase(storeRepo repository.StoreRepo, kafka map[string]*kafka.Writer) StoreUsecase {
	return &storeUsecase{storeRepo, kafka}
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

	writer, ok := u.kafka["products-request"]
	if !ok {
		return nil, utils.ErrNoTopic
	}

	payload := map[string]interface{}{
		"store_id":       storeId,
		"correlation_id": corrID,
	}
	value, _ := json.Marshal(payload)

	msg := kafka.Message{
		Key:   []byte(corrID),
		Value: value,
	}

	if err := writer.WriteMessages(context.Background(), msg); err != nil {
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

	if err := u.SendNotifRequest(corrId, req.Email, message, "store", "create"); err != nil {
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

	if err := u.SendNotifRequest(corrId, req.Email, message, "store", "update"); err != nil {
		return err
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

	if err := u.SendNotifRequest(corrId, email, storeId, "store", "delete"); err != nil {
		return err
	}

	return nil

}

func (u *storeUsecase) SendValidationResponse(userId, storeId uint, correlationID string) error {
	writer, ok := u.kafka["store-validation-response"]
	if !ok {
		return fmt.Errorf("writer for topic  not found")
	}

	valid, err := u.storeRepo.IsUserAdminStore(userId, storeId)
	if err != nil {
		return err
	}
	payload := map[string]interface{}{
		"correlation_id": correlationID,
		"is_valid":       valid,
	}
	value, _ := json.Marshal(payload)

	msg := kafka.Message{
		Key:   []byte(correlationID),
		Value: value,
	}

	return writer.WriteMessages(context.Background(), msg)
}

func (u *storeUsecase) SendNotifRequest(correlationID, email string, message interface{}, service, action string) error {
	writer, ok := u.kafka["notification-request"]
	if !ok {
		return utils.ErrFailedKafkaWrite
	}

	data := map[string]interface{}{
		"correlation_id": correlationID,
		"email":          email,
		"service":        service,
		"action":         action,
		"message":        message,
	}

	value, _ := json.Marshal(data)
	msg := kafka.Message{
		Key:   []byte(correlationID),
		Value: value,
	}

	return writer.WriteMessages(context.Background(), msg)
}
