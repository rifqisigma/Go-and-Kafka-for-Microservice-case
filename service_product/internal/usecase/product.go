package usecase

import (
	"context"
	"encoding/json"
	"service_product/dto"
	"service_product/helper/utils"
	"service_product/internal/repository"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
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
	SendValidationStoreRequest(storeId, userId uint, corelationId string) error
	SendNotifRequest(correlationID, email string, message interface{}, service, action string) error
}

type productUsecase struct {
	productRepo repository.ProductRepo
	kafka       map[string]*kafka.Writer
}

func NewProductUsecase(productRepo repository.ProductRepo, kafka map[string]*kafka.Writer) ProductUsecase {
	return &productUsecase{productRepo, kafka}
}

func (u *productUsecase) CreateProduct(req *dto.CreateProductReq) error {

	corrID := uuid.NewString()

	if err := u.SendValidationStoreRequest(req.StoreID, req.UserID, corrID); err != nil {
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

	if err := u.SendNotifRequest(corrID, req.Email, message, "product", "create"); err != nil {
		return err
	}
	return nil
}

func (u *productUsecase) UpdateProduct(req *dto.UpdateProductReq) error {
	corrID := uuid.NewString()

	if err := u.SendValidationStoreRequest(req.StoreID, req.UserID, corrID); err != nil {
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
	if err := u.SendNotifRequest(corrID, req.Email, message, "product", "update"); err != nil {
		return err
	}
	return nil
}

func (u *productUsecase) DeleteProduct(userId, storeId, id uint, email string) error {
	corrID := uuid.NewString()

	if err := u.SendValidationStoreRequest(storeId, userId, corrID); err != nil {
		return err
	}

	var isValid bool
	if err := u.productRepo.WaitForResponse(corrID, &isValid); err != nil {
		return err
	}
	if !isValid {
		return utils.ErrNotAdmin
	}

	writer := u.kafka["validation-product-response"]

	payload := map[string]interface{}{
		"correlation_id": corrID,
		"deleted":        true,
		"stock":          0,
		"product_id":     id,
	}
	value, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(corrID),
		Value: value,
	}

	if err := writer.WriteMessages(context.Background(), msg); err != nil {
		return utils.ErrFailedKafkaWrite
	}

	if err := u.SendNotifRequest(corrID, email, id, "product", "create"); err != nil {
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

	writer := u.kafka["products-response"]

	payload := map[string]interface{}{
		"correlation_id": correlation_id,
		"data":           products,
	}
	value, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(correlation_id),
		Value: value,
	}

	return writer.WriteMessages(context.Background(), msg)
}

func (u *productUsecase) SendValidationCartResponse(productId uint, correlation_id string) error {
	result, err := u.productRepo.ProductValidation(productId)
	if err != nil {
		return err
	}

	writer := u.kafka["validation-product-response"]

	payload := map[string]interface{}{
		"correlation_id": correlation_id,
		"deleted":        result.Deleted,
		"stock":          result.Stock,
		"product_id":     result.ProductId,
	}
	value, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(correlation_id),
		Value: value,
	}

	return writer.WriteMessages(context.Background(), msg)
}

func (u *productUsecase) SendValidationStoreRequest(storeId, userId uint, corelationId string) error {
	writer, ok := u.kafka["store-validation-request"]
	if !ok {
		return utils.ErrNoTopic
	}

	payload := map[string]interface{}{
		"store_id":       storeId,
		"user_id":        userId,
		"correlation_id": corelationId,
	}
	value, _ := json.Marshal(payload)

	msg := kafka.Message{
		Key:   []byte(corelationId),
		Value: value,
	}

	if err := writer.WriteMessages(context.Background(), msg); err != nil {
		return utils.ErrFailedKafkaWrite
	}
	return nil
}

func (u *productUsecase) SendNotifRequest(correlationID, email string, message interface{}, service, action string) error {
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
