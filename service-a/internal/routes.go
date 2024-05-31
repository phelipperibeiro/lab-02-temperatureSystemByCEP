package internal

import (
	"net/http"

	"github.com/gorilla/mux"
)

func SetupRouter() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/cep", handleCep).Methods("POST")
	return router
}
