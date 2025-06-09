package route

import (
	"net/http"
	"service_store/helper/middleware"
	"service_store/internal/handler"

	"github.com/gorilla/mux"
)

func SetupRoute(store *handler.StoreHandler) *mux.Router {
	r := mux.NewRouter()

	useM := r.PathPrefix("/store").Subrouter()
	useM.Use(middleware.AuthMiddleware)

	useM.HandleFunc("/create", store.CreateStore).Methods(http.MethodPost)
	useM.HandleFunc("/update/{storeId}", store.UpdateStore).Methods(http.MethodPut)
	useM.HandleFunc("/delete/{storeId}", store.DeleteStore).Methods(http.MethodDelete)
	useM.HandleFunc("/get/{storeId}", store.GetMyStore).Methods(http.MethodGet)
	useM.HandleFunc("/get", store.GetAllStore).Methods(http.MethodGet)
	return r
}
