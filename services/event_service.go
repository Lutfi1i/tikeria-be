package services

import (
	"errors"
	"strings"
	"time"

	"tikeria-be/dto"
	"tikeria-be/models"
	"tikeria-be/repositories"
)

type EventService interface {
	CreateEvent(organizerID uint, req dto.CreateEventRequest, isPro ...bool) (*models.Event, error)
	UpdateEvent(organizerID uint, eventID uint, req dto.UpdateEventRequest) (*models.Event, error)
	GetAllEvents() ([]models.Event, error)
	GetEventByID(id uint) (*models.Event, error)
	GetEventsByOrganizer(organizerID uint) ([]models.Event, error)
	AddTicketCategory(organizerID uint, eventID uint, req dto.TicketInput, isPro bool) (*models.Ticket, error)
	UpdateTicketCategory(organizerID uint, eventID uint, categoryID uint, req dto.TicketInput, isPro bool) (*models.Ticket, error)
	DeleteTicketCategory(organizerID uint, eventID uint, categoryID uint) error
	SubmitEventApproval(organizerID uint, eventID uint) (*models.Event, error)
	ApproveEvent(adminIdentifier string, eventID uint) (*models.Event, error)
	RejectEvent(adminIdentifier string, eventID uint, rejectReason string) (*models.Event, error)
}

type eventService struct {
	eventRepo repositories.EventRepository
	userRepo  repositories.UserRepository
}

func NewEventService(eventRepo repositories.EventRepository, userRepos ...repositories.UserRepository) EventService {
	var uRepo repositories.UserRepository
	if len(userRepos) > 0 {
		uRepo = userRepos[0]
	}
	return &eventService{eventRepo: eventRepo, userRepo: uRepo}
}

func (s *eventService) isOrganizerPro(organizerID uint, explicitIsPro ...bool) bool {
	if len(explicitIsPro) > 0 && explicitIsPro[0] {
		return true
	}
	if s.userRepo != nil {
		u, err := s.userRepo.FindByID(organizerID)
		if err == nil && u != nil && strings.EqualFold(u.Package, "pro") {
			return true
		}
	}
	return false
}

func isEventLocked(status string) bool {
	st := strings.ToLower(strings.TrimSpace(status))
	return st == "waiting approval" || st == "published"
}

func (s *eventService) CreateEvent(organizerID uint, req dto.CreateEventRequest, isPro ...bool) (*models.Event, error) {
	pro := s.isOrganizerPro(organizerID, isPro...)

	if !pro {
		activeCount, err := s.eventRepo.CountActiveEventsByOrganizerID(organizerID)
		if err == nil && activeCount >= 1 {
			return nil, errors.New("PACKAGE_LIMIT_ACTIVE_EVENT: Basic package allows 1 active event. Upgrade to Pro or wait until the current event ends.")
		}
		if len(req.Tickets) > 1 {
			return nil, errors.New("PACKAGE_LIMIT_TICKET_CATEGORY: Basic package allows maximum 1 ticket category. Upgrade to Pro for unlimited categories.")
		}
	}

	totalQuota := 0
	var tickets []models.Ticket
	for _, t := range req.Tickets {
		catName := strings.TrimSpace(t.CategoryName)
		if catName == "" {
			catName = strings.TrimSpace(t.Name)
		}
		if catName == "" {
			return nil, errors.New("ticket category name is required")
		}
		maxTx := t.MaxTicketPerTransaction
		if maxTx <= 0 || maxTx > 5 {
			maxTx = 5
		}
		totalQuota += t.Quota
		tickets = append(tickets, models.Ticket{
			Name:                    catName,
			CategoryName:            catName,
			Description:             t.Description,
			Price:                   t.Price,
			Quota:                   t.Quota,
			Available:               t.Quota,
			MaxTicketPerTransaction: maxTx,
			IsActive:                true,
		})
	}

	if !pro && totalQuota > 1000 {
		return nil, errors.New("PACKAGE_LIMIT_QUOTA: Basic package allows max 1.000 tickets per event")
	}

	location := strings.TrimSpace(req.Location)
	if location == "" {
		location = strings.TrimSpace(req.VenueName)
	}
	category := strings.TrimSpace(req.Category)
	if category == "" {
		category = "Pensi"
	}
	template := strings.TrimSpace(req.EventTemplate)
	if template == "" {
		template = "pensi"
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "Draft"
	}

	event := models.Event{
		OrganizerID:   organizerID,
		Title:         req.Title,
		Description:   req.Description,
		Location:      location,
		City:          req.City,
		VenueName:     req.VenueName,
		VenueAddress:  req.VenueAddress,
		BannerURL:     req.BannerURL,
		Category:      category,
		EventTemplate: template,
		PICName:       req.PICName,
		PICWhatsApp:   req.PICWhatsApp,
		EventDate:     req.EventDate,
		Status:        status,
		Tickets:       tickets,
	}

	err := s.eventRepo.Create(&event)
	if err != nil {
		return nil, errors.New("failed to create new event")
	}

	return &event, nil
}

func (s *eventService) UpdateEvent(organizerID uint, eventID uint, req dto.UpdateEventRequest) (*models.Event, error) {
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}
	if event.OrganizerID != organizerID {
		return nil, errors.New("unauthorized: you do not own this event")
	}
	if isEventLocked(event.Status) {
		return nil, errors.New("EVENT_LOCKED: event cannot be edited while Waiting Approval or Published")
	}

	event.Title = req.Title
	event.Description = req.Description
	if req.Location != "" {
		event.Location = req.Location
	}
	if req.City != "" {
		event.City = req.City
	}
	if req.VenueName != "" {
		event.VenueName = req.VenueName
	}
	if req.VenueAddress != "" {
		event.VenueAddress = req.VenueAddress
	}
	if req.BannerURL != "" {
		event.BannerURL = req.BannerURL
	}
	if req.Category != "" {
		event.Category = req.Category
	}
	if req.EventTemplate != "" {
		event.EventTemplate = req.EventTemplate
	}
	if req.PICName != "" {
		event.PICName = req.PICName
	}
	if req.PICWhatsApp != "" {
		event.PICWhatsApp = req.PICWhatsApp
	}
	if !req.EventDate.IsZero() {
		event.EventDate = req.EventDate
	}

	if err := s.eventRepo.Update(event); err != nil {
		return nil, errors.New("failed to update event")
	}
	return event, nil
}

func (s *eventService) GetAllEvents() ([]models.Event, error) {
	return s.eventRepo.FindAll()
}

func (s *eventService) GetEventByID(id uint) (*models.Event, error) {
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("event not found")
	}
	return event, nil
}

func (s *eventService) GetEventsByOrganizer(organizerID uint) ([]models.Event, error) {
	return s.eventRepo.FindByOrganizerID(organizerID)
}

func (s *eventService) AddTicketCategory(organizerID uint, eventID uint, req dto.TicketInput, isPro bool) (*models.Ticket, error) {
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	if event.OrganizerID != organizerID {
		return nil, errors.New("unauthorized: you do not own this event")
	}

	if isEventLocked(event.Status) {
		return nil, errors.New("EVENT_LOCKED: cannot add category while event is Waiting Approval or Published")
	}

	catName := strings.TrimSpace(req.CategoryName)
	if catName == "" {
		catName = strings.TrimSpace(req.Name)
	}
	if catName == "" {
		return nil, errors.New("category_name or name is required")
	}

	pro := s.isOrganizerPro(organizerID, isPro)

	// Package enforcement: Basic tier allows maximum 1 ticket category and max 1000 quota
	if !pro {
		if len(event.Tickets) >= 1 {
			return nil, errors.New("PACKAGE_LIMIT_TICKET_CATEGORY: package limit reached: Basic package allows maximum 1 ticket category. Upgrade to Pro for unlimited categories.")
		}
		existingQuota := 0
		for _, t := range event.Tickets {
			existingQuota += t.Quota
		}
		if existingQuota+req.Quota > 1000 {
			return nil, errors.New("PACKAGE_LIMIT_QUOTA: Basic package allows max 1.000 tickets per event")
		}
	}

	maxTx := req.MaxTicketPerTransaction
	if maxTx <= 0 || maxTx > 5 {
		maxTx = 5
	}

	ticket := models.Ticket{
		EventID:                 eventID,
		Name:                    catName,
		CategoryName:            catName,
		Description:             req.Description,
		Price:                   req.Price,
		Quota:                   req.Quota,
		Available:               req.Quota,
		MaxTicketPerTransaction: maxTx,
		IsActive:                true,
	}

	if err := s.eventRepo.AddTicketCategory(&ticket); err != nil {
		return nil, errors.New("failed to add ticket category")
	}

	return &ticket, nil
}

func (s *eventService) UpdateTicketCategory(organizerID uint, eventID uint, categoryID uint, req dto.TicketInput, isPro bool) (*models.Ticket, error) {
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}
	if event.OrganizerID != organizerID {
		return nil, errors.New("unauthorized: you do not own this event")
	}
	if isEventLocked(event.Status) {
		return nil, errors.New("EVENT_LOCKED: cannot update category while event is Waiting Approval or Published")
	}

	ticket, err := s.eventRepo.FindTicketByID(categoryID)
	if err != nil || ticket.EventID != eventID {
		return nil, errors.New("ticket category not found")
	}

	pro := s.isOrganizerPro(organizerID, isPro)
	if !pro && req.Quota > 1000 {
		return nil, errors.New("PACKAGE_LIMIT_QUOTA: Basic package allows max 1.000 tickets per event")
	}

	catName := strings.TrimSpace(req.CategoryName)
	if catName == "" {
		catName = strings.TrimSpace(req.Name)
	}
	if catName != "" {
		ticket.Name = catName
		ticket.CategoryName = catName
	}
	if req.Description != "" {
		ticket.Description = req.Description
	}
	ticket.Price = req.Price
	ticket.Quota = req.Quota
	ticket.Available = req.Quota - ticket.SoldQuantity
	if req.MaxTicketPerTransaction > 0 && req.MaxTicketPerTransaction <= 5 {
		ticket.MaxTicketPerTransaction = req.MaxTicketPerTransaction
	}
	if req.IsActive != nil {
		ticket.IsActive = *req.IsActive
	}

	if err := s.eventRepo.UpdateTicket(ticket); err != nil {
		return nil, errors.New("failed to update ticket category")
	}
	return ticket, nil
}

func (s *eventService) DeleteTicketCategory(organizerID uint, eventID uint, categoryID uint) error {
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return errors.New("event not found")
	}
	if event.OrganizerID != organizerID {
		return errors.New("unauthorized: you do not own this event")
	}
	if isEventLocked(event.Status) {
		return errors.New("EVENT_LOCKED: cannot delete category while event is Waiting Approval or Published")
	}

	ticket, err := s.eventRepo.FindTicketByID(categoryID)
	if err != nil || ticket.EventID != eventID {
		return errors.New("ticket category not found")
	}
	if ticket.SoldQuantity > 0 {
		return errors.New("ticket category already has sold tickets; deactivate instead of deleting")
	}
	return s.eventRepo.DeleteTicket(categoryID)
}

func (s *eventService) SubmitEventApproval(organizerID uint, eventID uint) (*models.Event, error) {
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}
	if event.OrganizerID != organizerID {
		return nil, errors.New("unauthorized: you do not own this event")
	}
	st := strings.ToLower(strings.TrimSpace(event.Status))
	if st != "draft" && st != "reject" {
		return nil, errors.New("event can only be submitted for approval from Draft or Reject status")
	}
	if len(event.Tickets) == 0 {
		return nil, errors.New("event must have at least 1 ticket category before submitting for approval")
	}

	event.Status = "Waiting Approval"
	event.RejectReason = ""
	if err := s.eventRepo.Update(event); err != nil {
		return nil, errors.New("failed to submit event for approval")
	}
	return event, nil
}

func (s *eventService) ApproveEvent(adminIdentifier string, eventID uint) (*models.Event, error) {
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}
	if !strings.EqualFold(event.Status, "Waiting Approval") {
		return nil, errors.New("only events in Waiting Approval status can be approved")
	}

	now := time.Now()
	event.Status = "Published"
	event.ReviewedBy = adminIdentifier
	event.ReviewedAt = &now
	event.PublishedAt = &now
	if err := s.eventRepo.Update(event); err != nil {
		return nil, errors.New("failed to approve event")
	}
	return event, nil
}

func (s *eventService) RejectEvent(adminIdentifier string, eventID uint, rejectReason string) (*models.Event, error) {
	if strings.TrimSpace(rejectReason) == "" {
		return nil, errors.New("reject_reason is required when rejecting an event")
	}
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}
	if !strings.EqualFold(event.Status, "Waiting Approval") {
		return nil, errors.New("only events in Waiting Approval status can be rejected")
	}

	now := time.Now()
	event.Status = "Reject"
	event.RejectReason = rejectReason
	event.ReviewedBy = adminIdentifier
	event.ReviewedAt = &now
	if err := s.eventRepo.Update(event); err != nil {
		return nil, errors.New("failed to reject event")
	}
	return event, nil
}
