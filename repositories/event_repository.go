package repositories

import (
	"tikeria-be/models"

	"gorm.io/gorm"
)

type EventRepository interface {
	Create(event *models.Event) error
	FindAll() ([]models.Event, error)
	FindByID(id uint) (*models.Event, error)
	FindByOrganizerID(organizerID uint) ([]models.Event, error)
	FindTicketByID(ticketID uint) (*models.Ticket, error)
	UpdateTicket(ticket *models.Ticket) error
	AddTicketCategory(ticket *models.Ticket) error
}

type eventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) Create(event *models.Event) error {
	return r.db.Create(event).Error
}

func (r *eventRepository) FindAll() ([]models.Event, error) {
	var events []models.Event
	err := r.db.Preload("Organizer").Preload("Tickets").Find(&events).Error
	return events, err
}

func (r *eventRepository) FindByID(id uint) (*models.Event, error) {
	var event models.Event
	err := r.db.Preload("Organizer").Preload("Tickets").First(&event, id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *eventRepository) FindTicketByID(ticketID uint) (*models.Ticket, error) {
	var ticket models.Ticket
	err := r.db.First(&ticket, ticketID).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *eventRepository) UpdateTicket(ticket *models.Ticket) error {
	return r.db.Save(ticket).Error
}

func (r *eventRepository) FindByOrganizerID(organizerID uint) ([]models.Event, error) {
	var events []models.Event
	err := r.db.Preload("Tickets").Where("organizer_id = ?", organizerID).Find(&events).Error
	return events, err
}

func (r *eventRepository) AddTicketCategory(ticket *models.Ticket) error {
	return r.db.Create(ticket).Error
}
