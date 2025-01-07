package clients

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.crja72.ru/gospec/go8/payment/internal/models"
)

type MockRoundTripper2 struct {
	Response *http.Response
	Err      error
}

func (m *MockRoundTripper2) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.Response, m.Err
}

func createMockHTTPClient2(responseBody string, statusCode int, err error) *http.Client {
	mockTransport := &MockRoundTripper2{
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

func TestCheckPaymentStatus_Success(t *testing.T) {
	mockResponse := `{
		"operations": [{"status": "success"}],
		"error": ""
	}`

	mockClient := createMockHTTPClient2(mockResponse, http.StatusOK, nil)
	client := &YooMoneyClient{
		Client:     mockClient,
		Token:      "mock-token",
		ClientID:   "mock-client-id",
		APIBaseURL: "https://mock-yoomoney.ru",
	}

	status, err := client.CheckPaymentStatus("valid-label")
	assert.NoError(t, err)
	assert.Equal(t, "success", status)
}

func TestCheckPaymentStatus_Failure(t *testing.T) {
	mockResponse := `{
		"operations": [{"status": "refused"}],
		"error": ""
	}`

	mockClient := createMockHTTPClient2(mockResponse, http.StatusOK, nil)
	client := &YooMoneyClient{
		Client:     mockClient,
		Token:      "mock-token",
		ClientID:   "mock-client-id",
		APIBaseURL: "https://mock-yoomoney.ru",
	}

	status, err := client.CheckPaymentStatus("valid-label")
	assert.Error(t, err)
	assert.Equal(t, "failed", status)
	assert.Contains(t, err.Error(), "payment refused")
}

func TestCheckPaymentStatus_Error(t *testing.T) {
	mockResponse := `{"error": "some-error"}`
	mockClient := createMockHTTPClient2(mockResponse, http.StatusOK, nil)
	client := &YooMoneyClient{
		Client:     mockClient,
		Token:      "mock-token",
		ClientID:   "mock-client-id",
		APIBaseURL: "https://mock-yoomoney.ru",
	}

	status, err := client.CheckPaymentStatus("valid-label")
	assert.Error(t, err)
	assert.Equal(t, "error", status)
	assert.Contains(t, err.Error(), "API error")
}

func TestCreateTransfer_Success(t *testing.T) {
	mockResponse := `{"status": "success"}`
	mockClient := createMockHTTPClient2(mockResponse, http.StatusOK, nil)

	client := &YooMoneyClient{
		Client:     mockClient,
		Token:      "mock-token",
		ClientID:   "mock-client-id",
		APIBaseURL: "https://mock-yoomoney.ru",
	}

	payment := &models.Payment{
		ID:       "payment-id",
		Amount:   100.0,
		Currency: "RUB",
		ToUserID: "recipient-id",
	}

	status, err := client.CreateTransfer(payment, "receiver-id")
	assert.NoError(t, err)
	assert.Equal(t, "success", status)
}

func TestCreateTransfer_InvalidPayment(t *testing.T) {
	client := &YooMoneyClient{}
	_, err := client.CreateTransfer(nil, "receiver-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "payment information is required")
}

func TestCreateTransfer_Failure(t *testing.T) {
	mockResponse := `{"status": "refused", "error": "insufficient funds"}`
	mockClient := createMockHTTPClient2(mockResponse, http.StatusOK, nil)

	client := &YooMoneyClient{
		Client:     mockClient,
		Token:      "mock-token",
		ClientID:   "mock-client-id",
		APIBaseURL: "https://mock-yoomoney.ru",
	}

	payment := &models.Payment{
		ID:       "payment-id",
		Amount:   100.0,
		Currency: "RUB",
		ToUserID: "recipient-id",
	}

	status, err := client.CreateTransfer(payment, "receiver-id")
	assert.Error(t, err)
	assert.Equal(t, "failed", status)
	assert.Contains(t, err.Error(), "insufficient funds")
}

func TestQuickPayment_Success(t *testing.T) {
	mockClient := createMockHTTPClient2("", http.StatusOK, nil)

	client := &YooMoneyClient{
		Client:     mockClient,
		Token:      "mock-token",
		ClientID:   "mock-client-id",
		APIBaseURL: "https://mock-yoomoney.ru",
	}

	url, err := client.QuickPayment("receiver-id", "targets", "PC", 100.0, "comment", "label", "additional-comment", "https://success.url")
	assert.NoError(t, err)
	assert.Contains(t, url, "receiver=receiver-id")
	assert.Contains(t, url, "sum=100.00")
}

func TestQuickPayment_InvalidInput(t *testing.T) {
	client := &YooMoneyClient{}
	_, err := client.QuickPayment("", "targets", "PC", 0, "", "", "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "receiver is required")
}
