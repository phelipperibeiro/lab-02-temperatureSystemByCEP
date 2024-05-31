package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/phelipperibeiro/lab-02-temperatureSystemByCEP/service-b/internal"
)

func main() {

	router := internal.SetupRouter()

	server := &http.Server{
		Addr:    ":8181",
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on :8181: %v\n", err)
		}
	}()
	log.Println("Server is ready to handle requests at :8181")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}
	log.Println("Server stopped")
}

// go run cmd/main.go
// curl -X POST -d '{"cep":"04942000"}' http://localhost:8181/cep
