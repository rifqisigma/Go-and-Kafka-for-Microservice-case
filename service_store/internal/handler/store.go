package handler

import (
	"encoding/json"
	"net/http"
	"service_store/dto"
	"service_store/helper/middleware"
	"service_store/helper/utils"
	"service_store/internal/usecase"
	"strconv"

	"github.com/gorilla/mux"
)

type StoreHandler struct {
	storeUscase usecase.StoreUsecase
}

func NewStoreHandler(storeUsecase usecase.StoreUsecase) *StoreHandler {
	return &StoreHandler{storeUsecase}
}

// store
func (h *StoreHandler) GetMyStore(w http.ResponseWriter, r *http.Request) {
	claimsRaw := r.Context().Value(middleware.UserContextKey)
	_, ok := claimsRaw.(*utils.JWTCLAIMS)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "protected api")
		return
	}

	params := mux.Vars(r)
	paramsId, err := strconv.Atoi(params["storeId"])
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "params tidak ditemukan")
		return
	}

	response, err := h.storeUscase.GetMyStore(uint(paramsId))
	if err != nil {
		switch err {
		case utils.ErrNotAdmin:
			utils.WriteError(w, http.StatusUnauthorized, "bukan admin")
			return
		default:
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *StoreHandler) GetAllStore(w http.ResponseWriter, r *http.Request) {
	claimsRaw := r.Context().Value(middleware.UserContextKey)
	_, ok := claimsRaw.(*utils.JWTCLAIMS)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "protected api")
		return
	}

	response, err := h.storeUscase.GetAllStore()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *StoreHandler) CreateStore(w http.ResponseWriter, r *http.Request) {
	claimsRaw := r.Context().Value(middleware.UserContextKey)
	claims, ok := claimsRaw.(*utils.JWTCLAIMS)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "protected api")
		return
	}

	var req dto.CreateStoreReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}

	req.Email = claims.Email
	req.AdminID = claims.UserID
	if err := h.storeUscase.CreateStore(&req); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *StoreHandler) UpdateStore(w http.ResponseWriter, r *http.Request) {
	claimsRaw := r.Context().Value(middleware.UserContextKey)
	claims, ok := claimsRaw.(*utils.JWTCLAIMS)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "protected api")
		return
	}

	var req dto.UpdateStoreReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}

	params := mux.Vars(r)
	paramsStoreId, err := strconv.Atoi(params["storeId"])
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "params tidak ditemukan")
		return
	}

	req.Email = claims.Email
	req.ID = uint(paramsStoreId)
	req.UserID = claims.UserID
	if err := h.storeUscase.UpdateStore(&req); err != nil {
		switch err {
		case utils.ErrNotAdmin:
			utils.WriteError(w, http.StatusForbidden, "bukan admin")
			return
		default:
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *StoreHandler) DeleteStore(w http.ResponseWriter, r *http.Request) {
	claimsRaw := r.Context().Value(middleware.UserContextKey)
	claims, ok := claimsRaw.(*utils.JWTCLAIMS)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "protected api")
		return
	}

	params := mux.Vars(r)
	paramsStoreId, err := strconv.Atoi(params["storeId"])
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "params tidak ditemukan")
		return
	}

	if err := h.storeUscase.DeleteStore(uint(paramsStoreId), claims.UserID, claims.Email); err != nil {
		switch err {
		case utils.ErrNotAdmin:
			utils.WriteError(w, http.StatusForbidden, "bukan admin")
			return
		default:
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}
