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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func main() {

	/////////////////////////////////////////
	/// Configurando o exportador Zipkin  ///
	/////////////////////////////////////////

	// Criando um exportador Zipkin
	exporter, err := zipkin.New(
		"http://zipkin:9411/api/v2/spans",
		zipkin.WithLogger(log.Default()),
	)

	if err != nil {
		log.Fatalf("Falha ao criar o exportador Zipkin: %v", err)
	}

	// Configurando o provedor
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("service-a"),
	)

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(resource),
	)

	defer func() {
		if err := tracerProvider.Shutdown(context.Background()); err != nil {
			log.Fatalf("Falha ao desligar o provedor de traçado: %v", err)
		}
	}()

	otel.SetTracerProvider(tracerProvider)

	//////////////////////////////////////
	/// Initialização do servidor HTTP ///
	//////////////////////////////////////

	// Configurando o roteador
	router := internal.SetupRouter()

	// Configurando o servidor HTTP
	server := &http.Server{
		Addr:    ":8080", // Servidor escutando na porta 8080
		Handler: router,
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