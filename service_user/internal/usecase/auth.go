package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"service_user/dto"
	"service_user/helper/utils"
	"service_user/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/sony/gobreaker"
)

type AuthUsecase interface {
	Register(req *dto.RegisterReq) error
	Login(req *dto.LoginReq) (string, error)
}

type authUsecase struct {
	authRepo     repository.AuthRepo
	kafka        map[string]*kafka.Writer
	writeBreaker *gobreaker.CircuitBreaker
}

func NewAuthUsecase(authRepo repository.AuthRepo, kafka map[string]*kafka.Writer) AuthUsecase {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "ProducerBreaker",
		MaxRequests: 5,
		Interval:    30 * time.Second,
		Timeout:     5 * time.Second,
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Printf("[Circuit Breaker: %s] status berubah dari %s ➜ %s\n", name, from.String(), to.String())
		},
	})
	return &authUsecase{authRepo, kafka, cb}
}

func (u *authUsecase) WriteKafkaMessage(topic string, key string, payload interface{}) error {
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

	data := map[string]interface{}{
		"correlation_id": corrId,
		"email":          req.Email,
		"service":        "user",
		"action":         "register",
		"message":        req.Email,
	}
	if err := u.WriteKafkaMessage("notification-request", corrId, data); err != nil {
		return err
	}
	return nil

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
