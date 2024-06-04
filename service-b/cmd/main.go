package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/phelipperibeiro/lab-02-temperatureSystemByCEP/service-b/internal"
)

func main() {

	/////////////////////////////////////////////////////////////
	///// Configuração do Tracing OpenTelemetry (Collector) /////
	/////////////////////////////////////////////////////////////

	shutdown := internal.InitTracer()
	defer shutdown()

	//////////////////////////////////////
	/// Initialização do servidor HTTP ///
	//////////////////////////////////////

	// Configurando o roteador
	router := internal.SetupRouter()

	// Configurando o servidor HTTP
	server := &http.Server{
		Addr:         ":8181",
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Iniciando o servidor em uma goroutine (Graceful shutdown)
	go func() {
		log.Println("Iniciando o servidor na porta 8181...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Falha ao iniciar o servidor: %v", err)
		}
	}()

	// Aguardando sinal para desligar o servidor
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan

	// Aguardando as conexões ativas terminarem antes de encerrar o servidor
	log.Println("Desligando o servidor...")

	// Tempo de espera para as conexões ativas terminarem
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Falha ao desligar o servidor: %v", err)
	}

	log.Println("Servidor desligado com sucesso.")
}

// go run cmd/main.go
// curl -X POST -d '{"cep":"04942000"}' http://localhost:8181/cep
