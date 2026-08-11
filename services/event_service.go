package services

import (
	"errors"

	"tikeria-be/dto"
	"tikeria-be/models"
	"tikeria-be/repositories"
)

type EventService interface {
	CreateEvent(organizerID uint, req dto.CreateEventRequest) (*models.Event, error)
	GetAllEvents() ([]models.Event, error)
	GetEventByID(id uint) (*models.Event, error)
}

type eventService struct {
	eventRepo repositories.EventRepository
}

func NewEventService(eventRepo repositories.EventRepository) EventService {
	return &eventService{eventRepo: eventRepo}
}

func (s *eventService) CreateEvent(organizerID uint, req dto.CreateEventRequest) (*models.Event, error) {
	var tickets []models.Ticket
	for _, t := range req.Tickets {
		tickets = append(tickets, models.Ticket{
			Name:      t.Name,
			Price:     t.Price,
			Quota:     t.Quota,
			Available: t.Quota,
		})
	}

	event := models.Event{
		OrganizerID: organizerID,
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		BannerURL:   req.BannerURL,
		Category:    req.Category,
		EventDate:   req.EventDate,
		Status:      "published",
		Tickets:     tickets,
	}

	err := s.eventRepo.Create(&event)
	if err != nil {
		return nil, errors.New("failed to create new event")
	}

	return &event, nil
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