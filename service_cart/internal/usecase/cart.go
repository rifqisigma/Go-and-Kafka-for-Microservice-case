package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"service_cart/dto"
	"service_cart/helper/utils"
	"service_cart/internal/repository"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type CartUsecase interface {
	GetMyCartItems(userId uint) ([]dto.CartItem, error)
	CreateCartItem(req *dto.CreateCartItemReq) error
	UpdateAmountCartItem(req *dto.UpdateAmountCartItemReq) error
	UpdatePaidCartItem(req *dto.UpdatePaidCartItemReq) error
	DeleteCartItem(userId, id uint) error

	//kafka
	CreateValidationWrite(correlationID string, productId uint) error
	UpdateIsDeleteProduct(id uint) error
	SendNotifRequest(correlationID, email string, message interface{}, service, action string) error
}

type cartUsecase struct {
	cartRepo repository.CartRepo
	kafka    map[string]*kafka.Writer
}

func NewCartUsecase(cartRepo repository.CartRepo, kafka map[string]*kafka.Writer) CartUsecase {
	return &cartUsecase{cartRepo, kafka}
}

func (u *cartUsecase) GetMyCartItems(userId uint) ([]dto.CartItem, error) {
	return u.cartRepo.GetMyCartItems(userId)
}

func (u *cartUsecase) CreateCartItem(req *dto.CreateCartItemReq) error {
	corrId := uuid.NewString()
	if err := u.CreateValidationWrite(corrId, req.ProductID); err != nil {
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
	if err := u.CreateValidationWrite(corrId, req.ProductID); err != nil {
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
	if err := u.CreateValidationWrite(corrId, req.ProductID); err != nil {
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
	if err := u.SendNotifRequest(corrId, req.Email, message, "cart", "paid"); err != nil {
		return err
	}
	return u.cartRepo.UpdatePaidCartItem(req)
}

func (u *cartUsecase) DeleteCartItem(userId, id uint) error {
	return u.cartRepo.DeleteCartItem(userId, id)
}

func (u *cartUsecase) CreateValidationWrite(correlationID string, productId uint) error {
	writer, ok := u.kafka["validation-product-request"]
	if !ok {
		return fmt.Errorf("writer for topic  not found")
	}

	payload := map[string]interface{}{
		"correlation_id": correlationID,
		"product_id":     productId,
	}

	value, _ := json.Marshal(payload)
	msg := kafka.Message{
		Key:   []byte(correlationID),
		Value: value,
	}

	return writer.WriteMessages(context.Background(), msg)
}

func (u *cartUsecase) UpdateIsDeleteProduct(id uint) error {
	return u.cartRepo.UpdateIsDeleteProduct(id)
}

func (u *cartUsecase) SendNotifRequest(correlationID, email string, message interface{}, service, action string) error {
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
