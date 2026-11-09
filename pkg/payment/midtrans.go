package payment

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"tikeria-be/config"
	"tikeria-be/models"
)

type SnapRequest struct {
	TransactionDetails TransactionDetails `json:"transaction_details"`
	CustomerDetails    CustomerDetails    `json:"customer_details"`
	ItemDetails        []ItemDetail       `json:"item_details,omitempty"`
	Expiry             *Expiry            `json:"expiry,omitempty"`
}

type TransactionDetails struct {
	OrderID     string  `json:"order_id"`
	GrossAmount float64 `json:"gross_amount"`
}

type CustomerDetails struct {
	FirstName string `json:"first_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

type ItemDetail struct {
	ID       string  `json:"id"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
	Name     string  `json:"name"`
}

type Expiry struct {
	Unit     string `json:"unit"`
	Duration int    `json:"duration"`
}

type SnapResponse struct {
	Token         string   `json:"token"`
	RedirectURL   string   `json:"redirect_url"`
	ErrorMessages []string `json:"error_messages,omitempty"`
}

type MidtransNotification struct {
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	TransactionID     string `json:"transaction_id"`
	StatusMessage     string `json:"status_message"`
	StatusCode        string `json:"status_code"`
	SignatureKey      string `json:"signature_key"`
	PaymentType       string `json:"payment_type"`
	OrderID           string `json:"order_id"`
	GrossAmount       string `json:"gross_amount"`
	FraudStatus       string `json:"fraud_status"`
	Currency          string `json:"currency"`
}

type MidtransClient interface {
	CreateSnapTransaction(order *models.Order, participant *models.ParticipantDetail) (*SnapResponse, error)
}

type midtransClient struct {
	cfg *config.AppConfig
}

func NewMidtransClient(cfg *config.AppConfig) MidtransClient {
	if cfg == nil {
		cfg = config.LoadConfig()
	}
	return &midtransClient{cfg: cfg}
}

func (m *midtransClient) CreateSnapTransaction(order *models.Order, participant *models.ParticipantDetail) (*SnapResponse, error) {
	if m.cfg == nil || m.cfg.MidtransServerKey == "" {
		// Mock response if server key is not configured yet
		return &SnapResponse{
			Token:       fmt.Sprintf("mock-snap-token-%s", order.OrderCode),
			RedirectURL: fmt.Sprintf("https://app.sandbox.midtrans.com/snap/v2/vtweb/mock-%s", order.OrderCode),
		}, nil
	}

	snapURL := "https://app.sandbox.midtrans.com/snap/v1/transactions"
	if m.cfg.MidtransIsProd {
		snapURL = "https://app.midtrans.com/snap/v1/transactions"
	}

	custName := "Tikeria Buyer"
	custEmail := "buyer@tikeria.id"
	custPhone := "08123456789"
	if participant != nil {
		if participant.FullName != "" {
			custName = participant.FullName
		} else if participant.FullNameKTP != "" {
			custName = participant.FullNameKTP
		}
		if participant.Email != "" {
			custEmail = participant.Email
		}
		if participant.Phone != "" {
			custPhone = participant.Phone
		} else if participant.WhatsApp != "" {
			custPhone = participant.WhatsApp
		}
	}

	items := []ItemDetail{}
	for _, it := range order.Items {
		items = append(items, ItemDetail{
			ID:       fmt.Sprintf("TICK-%d", it.TicketID),
			Price:    it.Price,
			Quantity: it.Quantity,
			Name:     it.CategoryName,
		})
	}
	if order.ServiceFee > 0 {
		items = append(items, ItemDetail{
			ID:       "FEE-PLATFORM",
			Price:    order.ServiceFee,
			Quantity: 1,
			Name:     "Biaya Layanan Platform",
		})
	}

	snapReq := SnapRequest{
		TransactionDetails: TransactionDetails{
			OrderID:     order.OrderCode,
			GrossAmount: order.TotalAmount,
		},
		CustomerDetails: CustomerDetails{
			FirstName: custName,
			Email:     custEmail,
			Phone:     custPhone,
		},
		ItemDetails: items,
		Expiry: &Expiry{
			Unit:     "minute",
			Duration: 15,
		},
	}

	jsonBytes, err := json.Marshal(snapReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, snapURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}

	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(m.cfg.MidtransServerKey+":"))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Midtrans Snap API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("midtrans error (%d): %s", resp.StatusCode, string(body))
	}

	var snapRes SnapResponse
	if err := json.Unmarshal(body, &snapRes); err != nil {
		return nil, err
	}

	return &snapRes, nil
}
