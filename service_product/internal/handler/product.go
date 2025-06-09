package handler

import (
	"encoding/json"
	"net/http"
	"service_product/dto"
	"service_product/helper/middleware"
	"service_product/helper/utils"
	"service_product/internal/usecase"
	"strconv"

	"github.com/gorilla/mux"
)

type StoreHandler struct {
	shopUsecase usecase.ProductUsecase
}

func NewStoreHandler(shopUsecase usecase.ProductUsecase) *StoreHandler {
	return &StoreHandler{shopUsecase}
}

func (h *StoreHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	claimsRaw := r.Context().Value(middleware.UserContextKey)
	claims, ok := claimsRaw.(*utils.JWTCLAIMS)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "protected api")
		return
	}

	var req dto.CreateProductReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Stock < 1 {
		utils.WriteError(w, http.StatusBadRequest, "invalid stock")
		return
	}

	params := mux.Vars(r)
	paramsStoreId, _ := strconv.Atoi(params["storeId"])

	req.Email = claims.Email
	req.UserID = claims.UserID
	req.StoreID = uint(paramsStoreId)
	if err := h.shopUsecase.CreateProduct(&req); err != nil {
		switch err {
		case utils.ErrNotAdmin:
			utils.WriteError(w, http.StatusUnauthorized, "bukan admin")
			return
		default:
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, nil)

}

func (h *StoreHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	claimsRaw := r.Context().Value(middleware.UserContextKey)
	claims, ok := claimsRaw.(*utils.JWTCLAIMS)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "protected api")
		return
	}

	params := mux.Vars(r)
	paramsProductId, _ := strconv.Atoi(params["productId"])
	paramsStoreId, _ := strconv.Atoi(params["storeId"])

	var req dto.UpdateProductReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Stock < 1 {
		utils.WriteError(w, http.StatusBadRequest, "invalid stock")
		return
	}

	req.Email = claims.Email
	req.ID = uint(paramsProductId)
	req.StoreID = uint(paramsStoreId)
	req.UserID = claims.UserID
	if err := h.shopUsecase.UpdateProduct(&req); err != nil {
		switch err {
		case utils.ErrNotAdmin:
			utils.WriteError(w, http.StatusUnauthorized, "bukan admin")
			return
		default:
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *StoreHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
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
	paramsProductId, err := strconv.Atoi(params["productId"])
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "params tidak ditemukan")
		return
	}

	if err := h.shopUsecase.DeleteProduct(claims.UserID, uint(paramsStoreId), uint(paramsProductId), claims.Email); err != nil {
		switch err {
		case utils.ErrNotAdmin:
			utils.WriteError(w, http.StatusUnauthorized, "bukan admin")
			return
		default:
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *StoreHandler) GetAllProduct(w http.ResponseWriter, r *http.Request) {
	claimsRaw := r.Context().Value(middleware.UserContextKey)
	_, ok := claimsRaw.(*utils.JWTCLAIMS)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "protected api")
		return
	}
	response, err := h.shopUsecase.GetAllProduct()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *StoreHandler) GetThisProduct(w http.ResponseWriter, r *http.Request) {
	claimsRaw := r.Context().Value(middleware.UserContextKey)
	_, ok := claimsRaw.(*utils.JWTCLAIMS)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "protected api")
		return
	}

	params := mux.Vars(r)
	paramsProductId, err := strconv.Atoi(params["productId"])
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}

	response, err := h.shopUsecase.GetProduct(uint(paramsProductId))
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, response)
}
