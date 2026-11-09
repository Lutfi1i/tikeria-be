package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"tikeria-be/config"
	"tikeria-be/controllers"
	"tikeria-be/dto"
	"tikeria-be/models"
	"tikeria-be/pkg/mailer"
	"tikeria-be/repositories"
	"tikeria-be/routes"
	"tikeria-be/services"
	"tikeria-be/utils"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupCheckoutAndOTPRouter(t *testing.T) (*gin.Engine, *gorm.DB, string, uint, uint) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open sqlite in memory: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Event{},
		&models.Ticket{},
		&models.TicketHold{},
		&models.ParticipantDetail{},
		&models.Order{},
		&models.OrderItem{},
		&models.Payment{},
	); err != nil {
		t.Fatalf("failed to migrate db: %v", err)
	}

	cfg := config.LoadConfig()
	cfg.DefaultServiceFee = 5000

	userRepo := repositories.NewUserRepository(db)
	eventRepo := repositories.NewEventRepository(db)
	checkoutRepo := repositories.NewCheckoutRepository(db)

	authService := services.NewAuthService(userRepo, eventRepo)
	eventService := services.NewEventService(eventRepo, userRepo)
	checkoutService := services.NewCheckoutService(checkoutRepo, eventRepo, userRepo, cfg)

	authController := controllers.NewAuthController(authService)
	eventController := controllers.NewEventController(eventService)
	checkoutController := controllers.NewCheckoutController(checkoutService)

	r := gin.New()
	r.Use(gin.Recovery())
	routes.RegisterAPIRoutes(r, authController, eventController, checkoutController)

	// Create test buyer user
	buyer := models.User{
		Name:          "Budi Buyer",
		Email:         "buyer@tikeria.test",
		Password:      "hashed_pass",
		Role:          "buyer",
		EmailVerified: true,
	}
	db.Create(&buyer)
	buyerToken, _ := utils.GenerateJWT(buyer.ID, buyer.Email, buyer.Role)

	// Create published event
	event := models.Event{
		OrganizerID: buyer.ID,
		Title:       "Konser Musik Kampus 2026",
		Description: "Konser tahunan kampus",
		Location:    "Jakarta International Expo",
		VenueName:   "Hall A",
		City:        "Jakarta",
		EventDate:   time.Now().Add(7 * 24 * time.Hour),
		Status:      "Published",
	}
	db.Create(&event)

	// Create ticket category
	ticket := models.Ticket{
		EventID:                 event.ID,
		Name:                    "Festival Presale",
		Price:                   150000,
		Quota:                   10,
		Available:               10,
		MaxTicketPerTransaction: 5,
		IsActive:                true,
	}
	db.Create(&ticket)

	return r, db, buyerToken, event.ID, ticket.ID
}

func TestOTPEmailTemplateAndRendering(t *testing.T) {
	htmlBody, textBody, err := mailer.RenderOTPEmail(mailer.OTPTemplateData{
		RecipientName: "Ahmad Khatami",
		OTPCode:       "849201",
		ExpiryMinutes: 10,
		SupportEmail:  "support@tikeria.id",
	})
	if err != nil {
		t.Fatalf("failed to render OTP email: %v", err)
	}

	if !strings.Contains(htmlBody, "849201") {
		t.Errorf("HTML body missing OTP code")
	}
	if !strings.Contains(htmlBody, "Ahmad Khatami") {
		t.Errorf("HTML body missing recipient name")
	}
	if !strings.Contains(htmlBody, "10 menit") {
		t.Errorf("HTML body missing expiry minutes notice")
	}
	if !strings.Contains(textBody, "849201") {
		t.Errorf("plain text body missing OTP code")
	}
}

func TestAuthRegisterOTPVerificationAndResend(t *testing.T) {
	r, db, _, _, _ := setupCheckoutAndOTPRouter(t)

	// 1. Register new buyer
	regPayload := `{"name":"Siti Rahma","email":"siti@tikeria.test","password":"secretpassword","role":"buyer"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(regPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 on register, got %d: %s", w.Code, w.Body.String())
	}

	// Verify user exists in DB with OTP
	var user models.User
	if err := db.Where("email = ?", "siti@tikeria.test").First(&user).Error; err != nil {
		t.Fatalf("registered user not found in db: %v", err)
	}
	if len(user.OTPCode) != 6 {
		t.Fatalf("expected 6-digit OTP code, got %q", user.OTPCode)
	}
	if user.EmailVerified {
		t.Fatalf("expected email_verified=false initially")
	}
	if user.OTPExpiresAt == nil || time.Now().After(*user.OTPExpiresAt) {
		t.Fatalf("expected valid future OTP expiry")
	}

	// 2. Try verify with wrong OTP
	badVerifyPayload := `{"email":"siti@tikeria.test","otp":"000000"}`
	req = httptest.NewRequest(http.MethodPost, "/api/auth/verify-otp", bytes.NewBufferString(badVerifyPayload))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad OTP, got %d", w.Code)
	}

	// 3. Verify with correct OTP
	goodVerifyPayload := fmt.Sprintf(`{"email":"siti@tikeria.test","otp":"%s"}`, user.OTPCode)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/verify-otp", bytes.NewBufferString(goodVerifyPayload))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for good OTP, got %d: %s", w.Code, w.Body.String())
	}

	db.First(&user, user.ID)
	if !user.EmailVerified {
		t.Fatalf("expected user.EmailVerified to be true after verification")
	}
	if user.OTPCode != "" {
		t.Fatalf("expected user.OTPCode to be cleared after verification")
	}

	// 4. Test Resend OTP
	resendPayload := `{"email":"siti@tikeria.test"}`
	req = httptest.NewRequest(http.MethodPost, "/api/auth/resend-otp", bytes.NewBufferString(resendPayload))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on resend OTP, got %d: %s", w.Code, w.Body.String())
	}

	db.First(&user, user.ID)
	if len(user.OTPCode) != 6 {
		t.Fatalf("expected new 6-digit OTP code after resend, got %q", user.OTPCode)
	}
}

func TestCheckoutHoldSummaryOrderAndPaymentLifecycle(t *testing.T) {
	r, db, token, eventID, ticketID := setupCheckoutAndOTPRouter(t)

	// 1. Create Hold (hold 2 tickets)
	holdPayload := fmt.Sprintf(`{"event_id":%d,"ticket_category_id":%d,"quantity":2}`, eventID, ticketID)
	req := httptest.NewRequest(http.MethodPost, "/api/checkout/hold", bytes.NewBufferString(holdPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create hold failed with code %d: %s", w.Code, w.Body.String())
	}

	var holdResp struct {
		Success bool                   `json:"success"`
		Data    dto.CreateHoldResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &holdResp); err != nil {
		t.Fatalf("failed to parse hold response: %v", err)
	}
	holdCode := holdResp.Data.HoldID
	if holdCode == "" {
		t.Fatalf("expected non-empty hold_id")
	}

	// Verify inventory decremented
	var ticket models.Ticket
	db.First(&ticket, ticketID)
	if ticket.Available != 8 {
		t.Fatalf("expected ticket.Available=8 after holding 2, got %d", ticket.Available)
	}

	// 2. Try holding more than remaining (try hold 9 tickets when only 8 available) -> 409 Conflict
	excessivePayload := fmt.Sprintf(`{"event_id":%d,"ticket_category_id":%d,"quantity":9}`, eventID, ticketID)
	req = httptest.NewRequest(http.MethodPost, "/api/checkout/hold", bytes.NewBufferString(excessivePayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	// In service, max quantity per transaction is 5 or stock check -> should be rejected
	if w.Code != http.StatusBadRequest && w.Code != http.StatusConflict {
		t.Fatalf("expected 400 or 409 for invalid/excessive quantity, got %d", w.Code)
	}

	// 3. Save Participant Details
	participantPayload := fmt.Sprintf(`{
		"hold_id": "%s",
		"full_name_ktp": "Ahmad Khatami",
		"email": "khatami@example.com",
		"phone": "081234567890",
		"nik": "3174000000001234",
		"institution": "Universitas Indonesia"
	}`, holdCode)
	req = httptest.NewRequest(http.MethodPost, "/api/checkout/participant", bytes.NewBufferString(participantPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("save participant failed with code %d: %s", w.Code, w.Body.String())
	}

	var partResp struct {
		Success bool                        `json:"success"`
		Data    dto.SaveParticipantResponse `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &partResp)
	if partResp.Data.Participant.MaskedNIK != "3174********1234" {
		t.Errorf("expected masked NIK 3174********1234, got %s", partResp.Data.Participant.MaskedNIK)
	}

	// 4. Get Checkout Summary
	summaryPayload := fmt.Sprintf(`{"hold_id":"%s"}`, holdCode)
	req = httptest.NewRequest(http.MethodPost, "/api/checkout/summary", bytes.NewBufferString(summaryPayload))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("get summary failed with code %d: %s", w.Code, w.Body.String())
	}

	var summaryResp struct {
		Success bool                        `json:"success"`
		Data    dto.CheckoutSummaryResponse `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &summaryResp)
	if summaryResp.Data.Item.Subtotal != 300000 {
		t.Errorf("expected subtotal 300000, got %f", summaryResp.Data.Item.Subtotal)
	}
	if summaryResp.Data.ServiceFee != 5000 {
		t.Errorf("expected flat service fee 5000, got %f", summaryResp.Data.ServiceFee)
	}
	if summaryResp.Data.TotalAmount != 305000 {
		t.Errorf("expected total amount 305000, got %f", summaryResp.Data.TotalAmount)
	}

	// 5. Create Order
	orderPayload := fmt.Sprintf(`{"hold_id":"%s"}`, holdCode)
	req = httptest.NewRequest(http.MethodPost, "/api/checkout/create-order", bytes.NewBufferString(orderPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create order failed with code %d: %s", w.Code, w.Body.String())
	}

	var orderResp struct {
		Success bool                    `json:"success"`
		Data    dto.CreateOrderResponse `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &orderResp)
	orderCode := orderResp.Data.OrderID
	invoiceNum := orderResp.Data.InvoiceNumber
	if orderResp.Data.Status != "Pending Payment" {
		t.Errorf("expected order status 'Pending Payment', got %s", orderResp.Data.Status)
	}
	if orderResp.Data.Payment.Status != "Waiting" {
		t.Errorf("expected payment status 'Waiting', got %s", orderResp.Data.Payment.Status)
	}

	// 6. Get Order Status by Order Code and Invoice
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/checkout/orders/%s/status", orderCode), nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get order status failed with code %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/orders/%s", invoiceNum), nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get order by invoice failed with code %d: %s", w.Code, w.Body.String())
	}

	// 7. Simulate Payment Success
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/payments/%s/simulate-success", orderCode), nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("simulate payment failed with code %d: %s", w.Code, w.Body.String())
	}

	var simResp struct {
		Success bool                    `json:"success"`
		Data    dto.OrderStatusResponse `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &simResp)
	if simResp.Data.OrderStatus != "Paid" {
		t.Errorf("expected order status 'Paid' after payment, got %s", simResp.Data.OrderStatus)
	}
	if simResp.Data.PaymentStatus != "Success" {
		t.Errorf("expected payment status 'Success', got %s", simResp.Data.PaymentStatus)
	}

	// Verify ticket sold quantity incremented by 2
	db.First(&ticket, ticketID)
	if ticket.SoldQuantity != 2 {
		t.Errorf("expected ticket.SoldQuantity=2, got %d", ticket.SoldQuantity)
	}

	// 8. Test Hold Release
	// Create another hold for 1 ticket
	holdPayload2 := fmt.Sprintf(`{"event_id":%d,"ticket_category_id":%d,"quantity":1}`, eventID, ticketID)
	req = httptest.NewRequest(http.MethodPost, "/api/checkout/hold", bytes.NewBufferString(holdPayload2))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	_ = json.Unmarshal(w.Body.Bytes(), &holdResp)
	holdCode2 := holdResp.Data.HoldID

	db.First(&ticket, ticketID)
	if ticket.Available != 7 {
		t.Fatalf("expected available 7 after hold, got %d", ticket.Available)
	}

	// Release hold
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/checkout/hold/%s", holdCode2), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("release hold failed with code %d: %s", w.Code, w.Body.String())
	}

	db.First(&ticket, ticketID)
	if ticket.Available != 8 {
		t.Fatalf("expected available restored to 8 after release, got %d", ticket.Available)
	}
}
