package handler

import (
	"encoding/json"
	"net/http"
	"service_cart/dto"
	"service_cart/helper/middleware"
	"service_cart/helper/utils"
	"service_cart/internal/usecase"
	"strconv"

	"github.com/gorilla/mux"
)

type CartHandler struct {
	cartUsecase usecase.CartUsecase
}

func NewCartpHandler(cartUsecase usecase.CartUsecase) *CartHandler {
	return &CartHandler{cartUsecase}
}

func (h *CartHandler) GetMyCartItems(w http.ResponseWriter, r *http.Request) {
	claimsRaw := r.Context().Value(middleware.UserContextKey)
	claims, ok := claimsRaw.(*utils.JWTCLAIMS)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "protected api")
		return
	}

	response, err := h.cartUsecase.GetMyCartItems(claims.UserID)
	if err != nil {
		switch err {
		case utils.ErrUnavaible:
			utils.WriteError(w, http.StatusOK, "kau belum ada cart items")
			return
		default:
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *CartHandler) CreateCartItem(w http.ResponseWriter, r *http.Request) {
	claimsRaw := r.Context().Value(middleware.UserContextKey)
	claims, ok := claimsRaw.(*utils.JWTCLAIMS)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "protected api")
		return
	}

	params := mux.Vars(r)
	paramsProductId, _ := strconv.Atoi(params["productId"])

	var req dto.CreateCartItemReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}

	if req.PurchaseAmount < 1 {
		utils.WriteError(w, http.StatusBadRequest, "invalid stock")
		return
	}

	req.UserID = claims.UserID
	req.ProductID = uint(paramsProductId)
	if err := h.cartUsecase.CreateCartItem(&req); err != nil {
		switch err {
		case utils.ErrStocknotEnough:
			utils.WriteError(w, http.StatusBadRequest, "stock tak cukup")
			return
		default:
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *CartHandler) UpdateAmountCartItem(w http.ResponseWriter, r *http.Request) {
	claimsRaw := r.Context().Value(middleware.UserContextKey)
	claims, ok := claimsRaw.(*utils.JWTCLAIMS)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "protected api")
		return
	}

	params := mux.Vars(r)
	paramsCartItemId, _ := strconv.Atoi(params["cartItemId"])
	paramsProductId, _ := strconv.Atoi(params["productId"])

	var req dto.UpdateAmountCartItemReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid")
		return
	}

	if req.PurchaseAmount < 1 {
		utils.WriteError(w, http.StatusBadRequest, "invalid stock")
		return
	}

	req.UserID = claims.UserID
	req.ID = uint(paramsCartItemId)
	req.ProductID = uint(paramsProductId)
	if err := h.cartUsecase.UpdateAmountCartItem(&req); err != nil {
		switch err {
		case utils.ErrStocknotEnough:
			utils.WriteError(w, http.StatusBadRequest, "stock tak cukup")
			return
		default:
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *CartHandler) UpdatePaidCartItem(w http.ResponseWriter, r *http.Request) {
	claimsRaw := r.Context().Value(middleware.UserContextKey)
	claims, ok := claimsRaw.(*utils.JWTCLAIMS)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "protected api")
		return
	}

	params := mux.Vars(r)
	paramsId, _ := strconv.Atoi(params["cartItemId"])
	paramsProductId, _ := strconv.Atoi(params["productId"])

	var req dto.UpdatePaidCartItemReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid")
		return
	}
	if req.PurchaseAmount < 1 {
		utils.WriteError(w, http.StatusBadRequest, "invalid stock")
		return
	}

	req.UserID = claims.UserID
	req.ID = uint(paramsId)
	req.ProductID = uint(paramsProductId)
	if err := h.cartUsecase.UpdatePaidCartItem(&req); err != nil {
		switch err {
		case utils.ErrStocknotEnough:
			utils.WriteError(w, http.StatusBadRequest, "stock tak cukup")
			return
		default:
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

	}

	utils.WriteJSON(w, http.StatusOK, nil)

}

func (h *CartHandler) DeleteCartItem(w http.ResponseWriter, r *http.Request) {
	claimsRaw := r.Context().Value(middleware.UserContextKey)
	claims, ok := claimsRaw.(*utils.JWTCLAIMS)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "protected api")
		return
	}

	params := mux.Vars(r)
	paramsId, _ := strconv.Atoi(params["cartItemId"])

	if err := h.cartUsecase.DeleteCartItem(claims.UserID, uint(paramsId)); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}
