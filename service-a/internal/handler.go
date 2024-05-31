package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
)

type Weather struct {
	TempC float64 `json:"temp_c"`
	TempF float64 `json:"temp_f"`
	TempK float64 `json:"temp_k"`
}

func handleCep(responseWriter http.ResponseWriter, request *http.Request) {

	tracer := otel.Tracer("servico-a")
	ctx := request.Context()
	ctx, span := tracer.Start(ctx, "handleCep") // Iniciando o span
	defer span.End()

	var data struct {
		CEP string `json:"cep"`
	}

	if err := json.NewDecoder(request.Body).Decode(&data); err != nil {
		http.Error(responseWriter, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	if len(data.CEP) != 8 {
		http.Error(responseWriter, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	fmt.Printf("Requisitando serviço B %s\n", data.CEP)

	location, weather, err := sendRequestToServiceB(ctx, data.CEP)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	// mock response
	// location := "São Paulo"
	// weather := &Weather{
	// 	TempC: 28.5,
	// 	TempF: 83.3,
	// 	TempK: 301.6,
	// }

	response := struct {
		City  string  `json:"city"`
		TempC float64 `json:"temp_C"`
		TempF float64 `json:"temp_F"`
		TempK float64 `json:"temp_K"`
	}{
		City:  location,
		TempC: weather.TempC,
		TempF: weather.TempF,
		TempK: weather.TempK,
	}

	responseWriter.Header().Set("Content-Type", "application/json")

	json.NewEncoder(responseWriter).Encode(response)
}

func sendRequestToServiceB(ctx context.Context, cep string) (string, *Weather, error) {

	tracer := otel.Tracer("servico-a")
	ctx, span := tracer.Start(ctx, "sendRequestToServiceB") // Iniciando o span
	defer span.End()

	url := "http://service-b:8181/cep"
	payload := struct {
		CEP string `json:"cep"`
	}{CEP: cep}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request payload: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to send request to service B: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("service B returned non-200 status code: %d", resp.StatusCode)
	}

	var response struct {
		City  string  `json:"city"`
		TempC float64 `json:"temp_C"`
		TempF float64 `json:"temp_F"`
		TempK float64 `json:"temp_K"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", nil, fmt.Errorf("failed to decode response from service B: %v", err)
	}

	weather := &Weather{TempC: response.TempC, TempF: response.TempF, TempK: response.TempK}

	return response.City, weather, nil
}
