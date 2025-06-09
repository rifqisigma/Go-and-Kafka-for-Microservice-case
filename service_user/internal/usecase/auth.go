package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"service_user/dto"
	"service_user/helper/utils"
	"service_user/internal/repository"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type AuthUsecase interface {
	Register(req *dto.RegisterReq) error
	Login(req *dto.LoginReq) (string, error)
}

type authUsecase struct {
	authRepo repository.AuthRepo
	kafka    map[string]*kafka.Writer
}

func NewAuthUsecase(authRepo repository.AuthRepo, kafka map[string]*kafka.Writer) AuthUsecase {
	return &authUsecase{authRepo, kafka}
}

func (u *authUsecase) Register(req *dto.RegisterReq) error {
	valid := utils.IsValidEmail(req.Email)
	if !valid {
		return utils.ErrInvalidEmail
	}
	hashsed, err := utils.HashPasswrd(req.Password)
	if err != nil {
		return err
	}

	req.Password = hashsed
	if err := u.authRepo.Register(req); err != nil {
		return err
	}

	corrId := uuid.NewString()
	writer, ok := u.kafka["notification-request"]
	if !ok {
		return utils.ErrInvalidWriter
	}

	data := map[string]interface{}{
		"correlation_id": corrId,
		"email":          req.Email,
		"service":        "user",
		"action":         "register",
		"message":        req.Email,
	}

	value, _ := json.Marshal(data)
	msg := kafka.Message{
		Key:   []byte(corrId),
		Value: value,
	}
	return writer.WriteMessages(context.Background(), msg)
}

func (u *authUsecase) Login(req *dto.LoginReq) (string, error) {
	valid := utils.IsValidEmail(req.Email)
	if !valid {
		return "", utils.ErrInvalidEmail
	}
	user, err := u.authRepo.LoginEmail(req.Email)
	if err != nil {
		return "", err
	}

	if valid := utils.ComparePassword(user.Password, req.Password); !valid {
		return "", errors.New("email dan password tidak cocok")
	}

	jwt, err := utils.GenerateJWT(user.Email, user.ID)
	if err != nil {
		return "", err
	}

	return jwt, nil
}
