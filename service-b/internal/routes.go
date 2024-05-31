package internal

import (
	"net/http"

	"github.com/gorilla/mux"
)

func SetupRouter() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/cep", handleCep).Methods("POST")

	return r
}
