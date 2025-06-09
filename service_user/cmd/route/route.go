package route

import (
	"net/http"
	"service_user/internal/handler"

	"github.com/gorilla/mux"
)

func SetupRoute(user *handler.AuthHandler) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/login", user.Login).Methods(http.MethodPost)
	r.HandleFunc("/register", user.Register).Methods(http.MethodPost)

	return r
}
