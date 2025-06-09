package utils

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, statusCode int, message string) {
	response := map[string]string{
		"error": message,
	}

	WriteJSON(w, statusCode, response)
}
