package services

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"tikeria-be/config"
	"tikeria-be/dto"
	"tikeria-be/models"
	"tikeria-be/repositories"
	"tikeria-be/utils"
)

type CheckoutService interface {
	CreateHold(userID uint, req dto.CreateHoldRequest) (*dto.CreateHoldResponse, error)
	ReleaseHold(userID uint, holdIdentifier string) error
	SaveParticipant(userID uint, req dto.SaveParticipantRequest) (*dto.SaveParticipantResponse, error)
	GetSummary(holdIdentifier string) (*dto.CheckoutSummaryResponse, error)
	CreateOrder(userID uint, req dto.CreateOrderRequest) (*dto.CreateOrderResponse, error)
	GetOrderStatus(identifier string) (*dto.OrderStatusResponse, error)
	SimulatePaymentSuccess(identifier string) (*dto.OrderStatusResponse, error)
}

type checkoutService struct {
	checkoutRepo repositories.CheckoutRepository
	eventRepo    repositories.EventRepository
	userRepo     repositories.UserRepository
	cfg          *config.AppConfig
}

func NewCheckoutService(
	checkoutRepo repositories.CheckoutRepository,
	eventRepo repositories.EventRepository,
	userRepo repositories.UserRepository,
	cfg *config.AppConfig,
) CheckoutService {
	if cfg == nil {
		cfg = config.LoadConfig()
	}
	return &checkoutService{
		checkoutRepo: checkoutRepo,
		eventRepo:    eventRepo,
		userRepo:     userRepo,
		cfg:          cfg,
	}
}

var nonDigitRegex = regexp.MustCompile(`\D+`)

func parseFlexibleID(val interface{}) (uint, error) {
	if val == nil {
		return 0, errors.New("id is required")
	}
	switch v := val.(type) {
	case float64:
		if v > 0 {
			return uint(v), nil
		}
	case int:
		if v > 0 {
			return uint(v), nil
		}
	case uint:
		if v > 0 {
			return v, nil
		}
	case string:
		trimmed := strings.TrimSpace(v)
		if id, err := strconv.Atoi(trimmed); err == nil && id > 0 {
			return uint(id), nil
		}
		// Extract trailing numbers, e.g. EVT-2026-00001 -> 1, CAT-2026-00005 -> 5
		parts := strings.Split(trimmed, "-")
		if len(parts) > 1 {
			last := parts[len(parts)-1]
			if id, err := strconv.Atoi(last); err == nil && id > 0 {
				return uint(id), nil
			}
		}
		digits := nonDigitRegex.ReplaceAllString(trimmed, "")
		if id, err := strconv.Atoi(digits); err == nil && id > 0 {
			return uint(id), nil
		}
	}
	return 0, fmt.Errorf("invalid identifier format: %v", val)
}

func (s *checkoutService) CreateHold(userID uint, req dto.CreateHoldRequest) (*dto.CreateHoldResponse, error) {
	eventID, err := parseFlexibleID(req.EventID)
	if err != nil {
		return nil, errors.New("valid event_id is required")
	}

	catRaw := req.TicketCategoryID
	if catRaw == nil {
		catRaw = req.TicketID
	}
	ticketID, err := parseFlexibleID(catRaw)
	if err != nil {
		return nil, errors.New("valid ticket_category_id or ticket_id is required")
	}

	event, err := s.eventRepo.FindByID(eventID)
	if err != nil || event == nil {
		return nil, errors.New("event not found")
	}

	// Validation: Event must be Published
	eventStatus := strings.ToLower(strings.TrimSpace(event.Status))
	if eventStatus != "published" {
		return nil, errors.New("event is not published yet")
	}

	ticket, err := s.eventRepo.FindTicketByID(ticketID)
	if err != nil || ticket == nil {
		return nil, errors.New("ticket category not found")
	}

	if req.Quantity < 1 {
		return nil, errors.New("quantity must be at least 1")
	}
	maxQty := 5
	if req.Quantity > maxQty {
		return nil, fmt.Errorf("maximum tickets per transaction is %d", maxQty)
	}

	// Validation: Available stock check
	if ticket.Available < req.Quantity {
		return nil, fmt.Errorf("ticket stock is not enough (available: %d)", ticket.Available)
	}

	now := time.Now()
	expiredAt := now.Add(15 * time.Minute)
	holdCode := fmt.Sprintf("HOLD-%d-%05d", now.Year(), now.UnixNano()%100000)

	hold := models.TicketHold{
		HoldCode:  holdCode,
		EventID:   event.ID,
		TicketID:  ticket.ID,
		UserID:    userID,
		Quantity:  req.Quantity,
		Status:    "Active",
		ExpiredAt: expiredAt,
	}

	if err := s.checkoutRepo.CreateHold(&hold); err != nil {
		return nil, fmt.Errorf("failed to create ticket hold: %w", err)
	}

	return &dto.CreateHoldResponse{
		HoldID:           hold.HoldCode,
		EventID:          fmt.Sprintf("EVT-%d-%05d", event.CreatedAt.Year(), event.ID),
		TicketCategoryID: fmt.Sprintf("CAT-%d-%05d", ticket.CreatedAt.Year(), ticket.ID),
		Quantity:         hold.Quantity,
		Status:           hold.Status,
		ExpiredAt:        dto.FormatTimeRFC3339(hold.ExpiredAt),
		RemainingSeconds: int(time.Until(hold.ExpiredAt).Seconds()),
	}, nil
}

func (s *checkoutService) ReleaseHold(userID uint, holdIdentifier string) error {
	hold, err := s.checkoutRepo.FindHoldByIDOrCode(holdIdentifier)
	if err != nil || hold == nil {
		return errors.New("ticket hold not found")
	}

	if hold.Status != "Active" {
		return fmt.Errorf("ticket hold is already %s", hold.Status)
	}

	return s.checkoutRepo.ReleaseHold(hold)
}

func (s *checkoutService) SaveParticipant(userID uint, req dto.SaveParticipantRequest) (*dto.SaveParticipantResponse, error) {
	holdIDStr := fmt.Sprintf("%v", req.HoldID)
	hold, err := s.checkoutRepo.FindHoldByIDOrCode(holdIDStr)
	if err != nil || hold == nil {
		return nil, errors.New("ticket hold not found")
	}

	if hold.Status != "Active" {
		return nil, fmt.Errorf("cannot add participant: hold is %s", hold.Status)
	}
	if time.Now().After(hold.ExpiredAt) {
		return nil, errors.New("ticket hold has expired")
	}

	fullName := strings.TrimSpace(req.FullName)
	if fullName == "" {
		fullName = strings.TrimSpace(req.FullNameKTP)
	}
	if fullName == "" {
		return nil, errors.New("full_name is required")
	}

	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		phone = strings.TrimSpace(req.WhatsApp)
	}
	if phone == "" {
		return nil, errors.New("phone or whatsapp is required")
	}

	holdIDVal := hold.ID
	participant := models.ParticipantDetail{
		HoldID:      &holdIDVal,
		FullName:    fullName,
		Email:       strings.TrimSpace(req.Email),
		Phone:       phone,
		WhatsApp:    phone,
		NIK:         strings.TrimSpace(req.NIK),
		BirthDate:   strings.TrimSpace(req.BirthDate),
		Gender:      strings.TrimSpace(req.Gender),
		Institution: strings.TrimSpace(req.Institution),
	}

	if err := s.checkoutRepo.SaveParticipant(&participant); err != nil {
		return nil, fmt.Errorf("failed to save participant detail: %w", err)
	}

	maskedNIK := ""
	if participant.NIK != "" {
		maskedNIK = utils.MaskNIK(participant.NIK)
	}

	return &dto.SaveParticipantResponse{
		HoldID: hold.HoldCode,
		Participant: dto.ParticipantDetailData{
			FullName:    participant.FullName,
			FullNameKTP: participant.FullName,
			Email:       participant.Email,
			Phone:       participant.Phone,
			WhatsApp:    participant.WhatsApp,
			MaskedNIK:   maskedNIK,
			BirthDate:   participant.BirthDate,
			Gender:      participant.Gender,
			Institution: participant.Institution,
		},
	}, nil
}

func (s *checkoutService) GetSummary(holdIdentifier string) (*dto.CheckoutSummaryResponse, error) {
	hold, err := s.checkoutRepo.FindHoldByIDOrCode(holdIdentifier)
	if err != nil || hold == nil {
		return nil, errors.New("ticket hold not found")
	}

	if hold.Status != "Active" {
		return nil, fmt.Errorf("hold status is %s", hold.Status)
	}
	if time.Now().After(hold.ExpiredAt) {
		return nil, errors.New("ticket hold has expired")
	}

	event, err := s.eventRepo.FindByID(hold.EventID)
	if err != nil || event == nil {
		return nil, errors.New("event not found for hold")
	}

	ticket, err := s.eventRepo.FindTicketByID(hold.TicketID)
	if err != nil || ticket == nil {
		return nil, errors.New("ticket category not found for hold")
	}

	subtotal := ticket.Price * float64(hold.Quantity)
	var serviceFee float64
	if subtotal > 0 {
		serviceFee = s.cfg.DefaultServiceFee
	}
	totalAmount := subtotal + serviceFee

	eventDate := event.EventDate.Format("2006-01-02")
	eventTime := event.EventDate.Format("15:04:05")

	venueName := event.VenueName
	if venueName == "" {
		venueName = event.Location
	}
	cityName := event.City
	if cityName == "" {
		cityName = event.Location
	}

	return &dto.CheckoutSummaryResponse{
		HoldID: hold.HoldCode,
		Event: dto.CheckoutEventSummary{
			EventID:   fmt.Sprintf("EVT-%d-%05d", event.CreatedAt.Year(), event.ID),
			Title:     event.Title,
			EventDate: eventDate,
			EventTime: eventTime,
			VenueName: venueName,
			City:      cityName,
		},
		Item: dto.CheckoutItemSummary{
			TicketCategoryID: fmt.Sprintf("CAT-%d-%05d", ticket.CreatedAt.Year(), ticket.ID),
			CategoryName:     ticket.Name,
			Price:            ticket.Price,
			Quantity:         hold.Quantity,
			Subtotal:         subtotal,
		},
		ServiceFee:  serviceFee,
		TotalAmount: totalAmount,
		ExpiredAt:   dto.FormatTimeRFC3339(hold.ExpiredAt),
	}, nil
}

func (s *checkoutService) CreateOrder(userID uint, req dto.CreateOrderRequest) (*dto.CreateOrderResponse, error) {
	holdIDStr := fmt.Sprintf("%v", req.HoldID)
	hold, err := s.checkoutRepo.FindHoldByIDOrCode(holdIDStr)
	if err != nil || hold == nil {
		return nil, errors.New("ticket hold not found")
	}

	if hold.Status != "Active" {
		return nil, fmt.Errorf("hold is %s", hold.Status)
	}
	if time.Now().After(hold.ExpiredAt) {
		return nil, errors.New("ticket hold has expired")
	}

	participant, err := s.checkoutRepo.FindParticipantByHoldID(hold.ID)
	if err != nil || participant == nil {
		return nil, errors.New("participant details must be completed before creating order")
	}

	ticket, err := s.eventRepo.FindTicketByID(hold.TicketID)
	if err != nil || ticket == nil {
		return nil, errors.New("ticket category not found")
	}

	now := time.Now()
	orderCode := fmt.Sprintf("ORD-%d-%05d", now.Year(), now.UnixNano()%100000)
	invoiceNumber := fmt.Sprintf("INV-%s-%05d", now.Format("20060102"), now.UnixNano()%100000)
	paymentCode := fmt.Sprintf("PAY-%d-%05d", now.Year(), now.UnixNano()%100000)

	subtotal := ticket.Price * float64(hold.Quantity)
	var serviceFee float64
	if subtotal > 0 {
		serviceFee = s.cfg.DefaultServiceFee
	}
	totalAmount := subtotal + serviceFee

	expiredAt := hold.ExpiredAt

	order := models.Order{
		OrderCode:     orderCode,
		InvoiceNumber: invoiceNumber,
		HoldID:        hold.ID,
		UserID:        userID,
		EventID:       hold.EventID,
		TicketID:      ticket.ID,
		Quantity:      hold.Quantity,
		Subtotal:      subtotal,
		ServiceFee:    serviceFee,
		TotalPrice:    totalAmount,
		TotalAmount:   totalAmount,
		Status:        "Pending Payment",
		ExpiredAt:     &expiredAt,
	}

	item := models.OrderItem{
		TicketID:     ticket.ID,
		CategoryName: ticket.Name,
		Price:        ticket.Price,
		Quantity:     hold.Quantity,
		Subtotal:     subtotal,
	}

	payment := models.Payment{
		PaymentCode:   paymentCode,
		Status:        "Waiting",
		Amount:        totalAmount,
		PaymentMethod: "Simulation",
	}

	if err := s.checkoutRepo.CreateOrderWithItemAndPayment(&order, &item, &payment); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return &dto.CreateOrderResponse{
		OrderID:       order.OrderCode,
		InvoiceNumber: order.InvoiceNumber,
		Status:        order.Status,
		Subtotal:      order.Subtotal,
		ServiceFee:    order.ServiceFee,
		TotalAmount:   order.TotalAmount,
		Payment: dto.PaymentSummary{
			PaymentID:     payment.PaymentCode,
			Status:        payment.Status,
			Amount:        payment.Amount,
			PaymentMethod: payment.PaymentMethod,
		},
		ExpiredAt: dto.FormatTimeRFC3339(expiredAt),
	}, nil
}

func (s *checkoutService) GetOrderStatus(identifier string) (*dto.OrderStatusResponse, error) {
	order, err := s.checkoutRepo.FindOrderByIDOrCodeOrInvoice(identifier)
	if err != nil || order == nil {
		return nil, errors.New("order not found")
	}

	paymentStatus := "Waiting"
	if order.Payment != nil {
		paymentStatus = order.Payment.Status
	}

	expStr := ""
	if order.ExpiredAt != nil {
		expStr = dto.FormatTimeRFC3339(*order.ExpiredAt)
	}

	return &dto.OrderStatusResponse{
		OrderID:       order.OrderCode,
		InvoiceNumber: order.InvoiceNumber,
		OrderStatus:   order.Status,
		PaymentStatus: paymentStatus,
		TotalAmount:   order.TotalAmount,
		ExpiredAt:     expStr,
	}, nil
}

func (s *checkoutService) SimulatePaymentSuccess(identifier string) (*dto.OrderStatusResponse, error) {
	order, err := s.checkoutRepo.FindOrderByIDOrCodeOrInvoice(identifier)
	if err != nil || order == nil {
		return nil, errors.New("order not found")
	}

	now := time.Now()
	order.Status = "Paid"
	order.PaidAt = &now
	_ = s.checkoutRepo.UpdateOrder(order)

	if order.Payment != nil {
		order.Payment.Status = "Success"
		order.Payment.PaidAt = &now
		_ = s.checkoutRepo.UpdatePayment(order.Payment)
	}

	// Update Ticket sold count
	ticket, err := s.eventRepo.FindTicketByID(order.TicketID)
	if err == nil && ticket != nil {
		ticket.SoldQuantity += order.Quantity
		_ = s.eventRepo.UpdateTicket(ticket)
	}

	return s.GetOrderStatus(identifier)
}
