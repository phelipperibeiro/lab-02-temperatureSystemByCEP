package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/phelipperibeiro/lab-02-temperatureSystemByCEP/service-a/internal"
)

func main() {

	// Configurando o roteador usando gorilla/mux
	router := internal.SetupRouter()

	// Configurando o servidor HTTP
	server := &http.Server{
		Addr:         ":8080", // Servidor escutando na porta 8080
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Iniciando o servidor em uma goroutine
	go func() {
		log.Println("Iniciando o servidor na porta 8080...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Falha ao iniciar o servidor: %v", err)
		}
	}()

	// Aguardando sinal para desligar o servidor
	signalChan := make(chan os.Signal, 1)                    // Canal para receber sinais
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM) // Captura os sinais de interrupção e termino do sistema
	<-signalChan                                             // Bloqueia até receber um sinal

	// Aguardando as conexões ativas terminarem antes de encerrar o servidor
	log.Println("Desligando o servidor...")

	// Tempo de espera para as conexões ativas terminarem
	const timeout = 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Falha ao desligar o servidor: %v", err)
	}

	log.Println("Servidor desligado com sucesso.")
}

// go run cmd/main.go
// curl -X POST -d '{"cep":"04942000"}' http://localhost:8080/cep -H "Content-Type: application/json"
