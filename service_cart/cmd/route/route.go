package route

import (
	"net/http"
	"service_cart/helper/middleware"
	"service_cart/internal/handler"

	"github.com/gorilla/mux"
)

func SetupRoute(cart *handler.CartHandler) *mux.Router {
	r := mux.NewRouter()

	useM := r.PathPrefix("/cart").Subrouter()
	useM.Use(middleware.AuthMiddleware)

	useM.HandleFunc("/create/{productId}", cart.CreateCartItem).Methods(http.MethodPost)
	useM.HandleFunc("/update-amount/{cartItemId}/{productId}", cart.UpdateAmountCartItem).Methods(http.MethodPut)
	useM.HandleFunc("/update-paid/{cartItemId}/{productId}", cart.UpdatePaidCartItem).Methods(http.MethodPut)
	useM.HandleFunc("/delete/{cartItemId}", cart.DeleteCartItem).Methods(http.MethodDelete)
	useM.HandleFunc("/me", cart.GetMyCartItems).Methods(http.MethodGet)

	return r
}
