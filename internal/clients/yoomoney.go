package clients

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gitlab.crja72.ru/gospec/go8/payment/internal/config"
	"gitlab.crja72.ru/gospec/go8/payment/internal/models"
)

type YooMoneyClient struct {
	Client     *http.Client
	Token      string
	ClientID   string
	APIBaseURL string
}

func NewYooMoneyClient(cfg *config.Config) *YooMoneyClient {
	return &YooMoneyClient{
		Client:     &http.Client{Timeout: 10 * time.Second},
		Token:      cfg.Yoomoney.Token,
		ClientID:   cfg.Yoomoney.ClientID,
		APIBaseURL: "https://yoomoney.ru",
	}
}

// CheckPaymentStatus проверяет статус платежа
func (c *YooMoneyClient) CheckPaymentStatus(label string) (string, error) {
	apiURL := fmt.Sprintf("%s/api/operation-history", c.APIBaseURL)

	params := url.Values{}
	params.Add("label", label)
	params.Add("records", "1")
	params.Add("type", "deposition")

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(params.Encode()))
	if err != nil {
		return "error", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.Client.Do(req)
	if err != nil {
		return "error", fmt.Errorf("failed to make API request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "error", fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "error", fmt.Errorf("API response status: %s, body: %s", resp.Status, string(body))
	}

	var response struct {
		Error      string `json:"error"`
		Operations []struct {
			Status string `json:"status"`
		} `json:"operations"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "error", fmt.Errorf("failed to decode response: %v", err)
	}

	if response.Error != "" {
		return "error", fmt.Errorf("API error: %s", response.Error)
	}

	if len(response.Operations) == 0 {
		return "error", fmt.Errorf("no operations found for label: %s", label)
	}

	status := response.Operations[0].Status

	switch status {
	case "success":
		return "success", nil
	case "refused":
		return "failed", fmt.Errorf("payment refused")
	case "in_progress":
		return "pending", nil
	default:
		return "error", fmt.Errorf("unexpected payment status: %s", status)
	}
}

// CreateTransfer Создает перевод
func (c *YooMoneyClient) CreateTransfer(payment *models.Payment, receiver string) (string, error) {
	if payment == nil {
		return "", fmt.Errorf("payment information is required")
	}
	if payment.ToUserID == "" {
		return "", fmt.Errorf("recipient (to_user_id) is required")
	}
	if payment.Amount <= 0 {
		return "", fmt.Errorf("amount must be greater than zero")
	}
	if payment.Currency == "" {
		return "", fmt.Errorf("currency is required")
	}
	if payment.ID == "" {
		return "", fmt.Errorf("payment ID is required")
	}

	apiURL := fmt.Sprintf("%s/api/request-payment", c.APIBaseURL)

	params := url.Values{}
	params.Add("pattern_id", "p2p")
	params.Add("to", receiver)
	params.Add("amount", strconv.FormatFloat(payment.Amount, 'f', 2, 64))
	params.Add("comment", payment.ID)
	params.Add("message", payment.ID)
	params.Add("label", payment.ID)
	params.Add("currency", payment.Currency)

	reqBody := params.Encode()

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make API request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API response status: %s, body: %s", resp.Status, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	status, ok := result["status"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response structure: missing 'status' field")
	}

	switch status {
	case "success":
		return "success", nil
	case "refused":
		errorMsg, _ := result["error"].(string)
		return "failed", fmt.Errorf("transfer refused: %s", errorMsg)
	default:
		return "error", fmt.Errorf("unexpected transfer status: %s", status)
	}
}

// QuickPayment создает ссылку для оплаты
func (c *YooMoneyClient) QuickPayment(receiver, targets, paymentType string, sum float64, formcomment, label, comment, successURL string) (string, error) {
	if receiver == "" {
		return "", fmt.Errorf("receiver is required")
	}
	if sum <= 0 {
		return "", fmt.Errorf("sum must be greater than zero")
	}

	baseURL := "https://yoomoney.ru/quickpay/confirm?"

	payload := url.Values{}
	payload.Add("receiver", receiver) // Receiver's YooMoney wallet
	payload.Add("quickpay-form", "shop")
	payload.Add("paymentType", paymentType) // Payment method: 'PC' or 'AC'
	payload.Add("sum", strconv.FormatFloat(sum, 'f', 2, 64))
	payload.Add("targets", targets) // Payment purpose/target description

	if formcomment != "" {
		payload.Add("formcomment", formcomment)
	}
	if label != "" {
		payload.Add("label", label) // Payment label
	}
	if comment != "" {
		payload.Add("comment", comment)
	}
	if successURL != "" {
		payload.Add("successURL", successURL) // Redirect URL after successful payment
	}

	finalURL := baseURL + payload.Encode()

	resp, err := http.Get(finalURL)
	if err != nil {
		return "", fmt.Errorf("failed to validate URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return finalURL, nil
	}

	return "", fmt.Errorf("validation failed with status: %s", resp.Status)
}
