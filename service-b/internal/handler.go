package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type Weather struct {
	TempC float64 `json:"temp_c"`
	TempF float64 `json:"temp_f"`
	TempK float64 `json:"temp_k"`
}

func handleCep(responseWriter http.ResponseWriter, request *http.Request) {
	trace := otel.Tracer("service-b")

	// Extrair o contexto do span da requisição HTTP
	carrier := propagation.HeaderCarrier(request.Header)
	ctx := request.Context()                                // white context
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier) // Get Header data from request and inject in context

	// Iniciar um novo span com o span do serviço A como parent
	ctx, span := trace.Start(ctx, "handleCep") // Iniciar um novo span
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

	fmt.Printf("consultando API de CEP para %s\n", data.CEP)

	// Consultar API de CEP
	location, err := getLocation(ctx, data.CEP)
	if err != nil {
		http.Error(responseWriter, "can not find zipcode", http.StatusNotFound)
		return
	}

	fmt.Printf("consultando API de clima para %s\n", location)

	// Consultar API de Clima
	weather, err := getWeather(ctx, location)
	if err != nil {
		http.Error(responseWriter, "error fetching weather", http.StatusInternalServerError)
		return
	}

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

func buildQuery(params map[string]string) string {
	var parts []string
	for key, value := range params {
		parts = append(parts, fmt.Sprintf("%s=%s", key, url.QueryEscape(value)))
	}
	return strings.Join(parts, "&")
}

func getLocation(ctx context.Context, cep string) (string, error) {

	tr := otel.Tracer("service-b")
	ctx, span := tr.Start(ctx, "getLocation")
	defer span.End()

	url := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned non-200 status code: %d", resp.StatusCode)
	}

	var response struct {
		Localidade string `json:"localidade"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return response.Localidade, nil
}

func getWeather(ctx context.Context, city string) (*Weather, error) {

	tr := otel.Tracer("service-b")
	ctx, span := tr.Start(ctx, "getWeather")
	defer span.End()

	apiKey := "776617dd5d694eaa94d33907242605" // token de acesso da API WeatherAPI
	params := map[string]string{
		"q":    city,
		"lang": "en",
		"key":  apiKey,
	}

	url := fmt.Sprintf("%s?%s", "http://api.weatherapi.com/v1/current.json", buildQuery(params))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status code: %d", resp.StatusCode)
	}

	var response struct {
		Current struct {
			TempC float64 `json:"temp_c"`
		} `json:"current"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	weather := &Weather{
		TempC: response.Current.TempC,
		TempF: response.Current.TempC*1.8 + 32,
		TempK: response.Current.TempC + 273.15,
	}

	return weather, nil
}
