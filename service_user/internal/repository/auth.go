package repository

import (
	"service_user/dto"
	"service_user/entity"

	"gorm.io/gorm"
)

type AuthRepo interface {
	Register(req *dto.RegisterReq) error
	LoginEmail(email string) (*entity.User, error)
}

type authRepo struct {
	db *gorm.DB
}

func NewAuthRepo(db *gorm.DB) AuthRepo {
	return &authRepo{db}
}

func (r *authRepo) Register(req *dto.RegisterReq) error {
	newUser := entity.User{
		Email:    req.Email,
		Password: req.Password,
		Username: req.Name,
	}

	return r.db.Model(&entity.User{}).Create(&newUser).Error
}

func (r *authRepo) LoginEmail(email string) (*entity.User, error) {
	var user entity.User
	if err := r.db.Model(&entity.User{}).Select("id", "email", "password").Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
