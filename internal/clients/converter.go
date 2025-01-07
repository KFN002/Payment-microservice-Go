package clients

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gitlab.crja72.ru/gospec/go8/payment/internal/config"
)

const baseURL = "https://api.fastforex.io/convert"

type ForexClient struct {
	APIKey string
	Client *http.Client
}

type ConversionResponse struct {
	Base   string             `json:"base"`
	Amount float64            `json:"amount"`
	Result map[string]float64 `json:"result"`
	Ms     int64              `json:"ms"`
}

func NewForexClient(cfg *config.Config) *ForexClient {
	return &ForexClient{
		APIKey: cfg.Forex.Key,
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

// ConvertCurrency Конвертер валют
func (fc *ForexClient) ConvertCurrency(from string, to string, amount float64) (float64, error) {
	url := fmt.Sprintf("%s?from=%s&to=%s&amount=%f&api_key=%s", baseURL, from, to, amount, fc.APIKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return -1, err
	}

	resp, err := fc.Client.Do(req)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1, err
	}

	if resp.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("error: received non-OK HTTP status %d", resp.StatusCode)
	}

	var response ConversionResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return -1, err
	}

	return response.Result[to], nil
}

// ConvertToRub конвертер валют в рубли
func (fc *ForexClient) ConvertToRub(amount float64, currency string) (float64, error) {
	return fc.ConvertCurrency(currency, "RUB", amount)
}
