package repositories

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"tikeria-be/models"

	"gorm.io/gorm"
)

type CheckoutRepository interface {
	CreateHold(hold *models.TicketHold) error
	FindHoldByIDOrCode(identifier string) (*models.TicketHold, error)
	UpdateHold(hold *models.TicketHold) error
	ReleaseHold(hold *models.TicketHold) error
	SaveParticipant(participant *models.ParticipantDetail) error
	FindParticipantByHoldID(holdID uint) (*models.ParticipantDetail, error)
	CreateOrderWithItemAndPayment(order *models.Order, item *models.OrderItem, payment *models.Payment) error
	FindOrderByIDOrCodeOrInvoice(identifier string) (*models.Order, error)
	UpdateOrder(order *models.Order) error
	UpdatePayment(payment *models.Payment) error
}

type checkoutRepository struct {
	db *gorm.DB
}

func NewCheckoutRepository(db *gorm.DB) CheckoutRepository {
	return &checkoutRepository{db: db}
}

func (r *checkoutRepository) CreateHold(hold *models.TicketHold) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Decrement available stock
		res := tx.Model(&models.Ticket{}).
			Where("id = ? AND available >= ?", hold.TicketID, hold.Quantity).
			UpdateColumn("available", gorm.Expr("available - ?", hold.Quantity))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("insufficient ticket stock")
		}

		if err := tx.Create(hold).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *checkoutRepository) FindHoldByIDOrCode(identifier string) (*models.TicketHold, error) {
	trimmed := strings.TrimSpace(identifier)
	var hold models.TicketHold

	// Check if integer ID
	if id, err := strconv.Atoi(trimmed); err == nil && id > 0 {
		if err := r.db.Preload("Participant").First(&hold, id).Error; err == nil {
			return &hold, nil
		}
	}

	// Try find by hold_code
	err := r.db.Preload("Participant").Where("hold_code = ?", trimmed).First(&hold).Error
	if err != nil {
		return nil, err
	}
	return &hold, nil
}

func (r *checkoutRepository) UpdateHold(hold *models.TicketHold) error {
	return r.db.Save(hold).Error
}

func (r *checkoutRepository) ReleaseHold(hold *models.TicketHold) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		res := tx.Model(&models.TicketHold{}).
			Where("id = ? AND status = ?", hold.ID, "Active").
			Updates(map[string]interface{}{
				"status":      "Released",
				"released_at": &now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected > 0 {
			if err := tx.Model(&models.Ticket{}).
				Where("id = ?", hold.TicketID).
				UpdateColumn("available", gorm.Expr("available + ?", hold.Quantity)).Error; err != nil {
				return err
			}
			hold.Status = "Released"
			hold.ReleasedAt = &now
		}
		return nil
	})
}

func (r *checkoutRepository) SaveParticipant(participant *models.ParticipantDetail) error {
	var existing models.ParticipantDetail
	if err := r.db.Where("hold_id = ?", participant.HoldID).First(&existing).Error; err == nil {
		participant.ID = existing.ID
		return r.db.Save(participant).Error
	}
	return r.db.Create(participant).Error
}

func (r *checkoutRepository) FindParticipantByHoldID(holdID uint) (*models.ParticipantDetail, error) {
	var p models.ParticipantDetail
	err := r.db.Where("hold_id = ?", holdID).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *checkoutRepository) CreateOrderWithItemAndPayment(order *models.Order, item *models.OrderItem, payment *models.Payment) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		item.OrderID = order.ID
		if err := tx.Create(item).Error; err != nil {
			return err
		}
		payment.OrderID = order.ID
		if err := tx.Create(payment).Error; err != nil {
			return err
		}

		// Mark hold as confirmed
		now := time.Now()
		_ = tx.Model(&models.TicketHold{}).
			Where("id = ?", order.HoldID).
			Updates(map[string]interface{}{
				"status":       "Confirmed",
				"confirmed_at": &now,
			})

		return nil
	})
}

func (r *checkoutRepository) FindOrderByIDOrCodeOrInvoice(identifier string) (*models.Order, error) {
	trimmed := strings.TrimSpace(identifier)
	var order models.Order

	query := r.db.Preload("Items").Preload("Payment").Preload("Event").Preload("Ticket")

	if id, err := strconv.Atoi(trimmed); err == nil && id > 0 {
		if err := query.First(&order, id).Error; err == nil {
			return &order, nil
		}
	}

	err := query.Where("order_code = ? OR invoice_number = ?", trimmed, trimmed).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *checkoutRepository) UpdateOrder(order *models.Order) error {
	return r.db.Save(order).Error
}

func (r *checkoutRepository) UpdatePayment(payment *models.Payment) error {
	return r.db.Save(payment).Error
}
