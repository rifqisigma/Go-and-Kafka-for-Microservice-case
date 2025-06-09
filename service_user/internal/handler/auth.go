package handler

import (
	"encoding/json"
	"net/http"
	"service_user/dto"
	"service_user/helper/utils"
	"service_user/internal/usecase"
)

type AuthHandler struct {
	authUsecase usecase.AuthUsecase
}

func NewAuthHandler(authUsecase usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}

	jwt, err := h.authUsecase.Login(&req)
	if err != nil {
		switch err {
		case utils.ErrInvalidEmail:
			utils.WriteError(w, http.StatusBadRequest, "invalid email")
			return
		default:
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"token": jwt,
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}

	if err := h.authUsecase.Register(&req); err != nil {
		switch err {
		case utils.ErrInvalidEmail:
			utils.WriteError(w, http.StatusBadRequest, "invalid email")
			return
		default:
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}
