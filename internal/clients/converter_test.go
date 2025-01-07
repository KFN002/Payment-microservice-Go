package clients

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.crja72.ru/gospec/go8/payment/internal/config"
)

type MockRoundTripper struct {
	Response *http.Response
	Err      error
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.Response, m.Err
}

func createMockHTTPClient(responseBody string, statusCode int, err error) *http.Client {
	mockTransport := &MockRoundTripper{
		Response: &http.Response{
			StatusCode: statusCode,
			Body:       ioutil.NopCloser(bytes.NewBufferString(responseBody)),
			Header:     make(http.Header),
		},
		Err: err,
	}
	return &http.Client{
		Transport: mockTransport,
		Timeout:   10 * time.Second,
	}
}

func TestConvertCurrency_ResponseCheck(t *testing.T) {
	mockResponse := `{
		"base": "USD",
		"amount": 100.0,
		"result": { "RUB": 7600.0 },
		"ms": 123
	}`

	cfg := &config.Config{
		Forex: config.Forex{Key: "mock-api-key"},
	}
	mockClient := createMockHTTPClient(mockResponse, http.StatusOK, nil)
	forexClient := &ForexClient{
		APIKey: cfg.Forex.Key,
		Client: mockClient,
	}

	_, err := forexClient.ConvertCurrency("USD", "RUB", 100)

	assert.NoError(t, err)
}

func TestConvertToRub_ResponseCheck(t *testing.T) {
	mockResponse := `{
		"base": "EUR",
		"amount": 50.0,
		"result": { "RUB": 3800.0 },
		"ms": 123
	}`

	cfg := &config.Config{
		Forex: config.Forex{Key: "mock-api-key"},
	}
	mockClient := createMockHTTPClient(mockResponse, http.StatusOK, nil)
	forexClient := &ForexClient{
		APIKey: cfg.Forex.Key,
		Client: mockClient,
	}

	_, err := forexClient.ConvertToRub(50, "EUR")

	assert.NoError(t, err)
}
