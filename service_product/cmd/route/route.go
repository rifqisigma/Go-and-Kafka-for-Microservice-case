package route

import (
	"net/http"
	"service_product/helper/middleware"
	"service_product/internal/handler"

	"github.com/gorilla/mux"
)

func SetupRoute(product *handler.StoreHandler) *mux.Router {
	r := mux.NewRouter()
	useM := r.PathPrefix("/product").Subrouter()
	useM.Use(middleware.AuthMiddleware)

	useM.HandleFunc("/create/{storeId}", product.CreateProduct).Methods(http.MethodPost)
	useM.HandleFunc("/update/{storeId}/{productId}", product.UpdateProduct).Methods(http.MethodPut)
	useM.HandleFunc("/delete/{storeId}/{productId}", product.DeleteProduct).Methods(http.MethodDelete)
	useM.HandleFunc("/getall", product.GetAllProduct).Methods(http.MethodGet)
	useM.HandleFunc("/get/{productId}", product.GetThisProduct).Methods(http.MethodGet)

	return r
}
