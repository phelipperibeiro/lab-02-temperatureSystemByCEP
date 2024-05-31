package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Weather struct {
	TempC float64 `json:"temp_c"`
	TempF float64 `json:"temp_f"`
	TempK float64 `json:"temp_k"`
}

func dd(data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("Erro ao serializar dados:", err)
		//os.Exit(1)
	}
	fmt.Println(string(jsonData))
}

func handleCep(w http.ResponseWriter, r *http.Request) {
	var request struct {
		CEP string `json:"cep"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	fmt.Printf("consultando API de CEP para %s\n", request.CEP)

	location, err := getLocation(request.CEP)
	if err != nil {
		http.Error(w, "can not find zipcode", http.StatusNotFound)
		return
	}

	fmt.Printf("consultando API de clima para %s\n", location)

	// Consultar API de Clima
	weather, err := getWeather(location)
	if err != nil {
		http.Error(w, "error fetching weather", http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getLocation(cep string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep))

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get location for %s", cep)
	}

	var data struct {
		Localidade string `json:"localidade"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	return data.Localidade, nil
}

func buildQuery(params map[string]string) string {
	var parts []string
	for key, value := range params {
		parts = append(parts, fmt.Sprintf("%s=%s", key, url.QueryEscape(value)))
	}
	return strings.Join(parts, "&")
}

func getWeather(city string) (*Weather, error) {

	apiKey := "776617dd5d694eaa94d33907242605" // token de acesso da API WeatherAPI

	params := map[string]string{
		"q":    city,
		"lang": "en",
		"key":  apiKey,
	}

	resp, err := http.Get(fmt.Sprintf("%s?%s", "http://api.weatherapi.com/v1/current.json", buildQuery(params)))

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get weather for %s", city)
	}

	var data struct {
		Current struct {
			TempC float64 `json:"temp_c"`
			TempF float64 `json:"temp_f"`
			TempK float64 `json:"temp_k"`
		} `json:"current"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	weather := &Weather{
		TempC: data.Current.TempC,
		TempF: data.Current.TempF,
		TempK: data.Current.TempK,
	}

	return weather, nil
}
