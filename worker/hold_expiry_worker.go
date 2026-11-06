package worker

import (
	"context"
	"log"
	"time"

	"tikeria-be/models"

	"gorm.io/gorm"
)

// ExpireOverdueHolds finds all Active TicketHolds whose ExpiredAt <= now,
// marks them as Expired, and restores their held quantity back to Ticket.Available.
func ExpireOverdueHolds(db *gorm.DB) (int64, error) {
	now := time.Now()
	var expiredHolds []models.TicketHold

	err := db.Where("status = ? AND expired_at <= ?", "Active", now).Find(&expiredHolds).Error
	if err != nil {
		return 0, err
	}
	if len(expiredHolds) == 0 {
		return 0, nil
	}

	var count int64
	err = db.Transaction(func(tx *gorm.DB) error {
		for _, hold := range expiredHolds {
			res := tx.Model(&models.TicketHold{}).
				Where("id = ? AND status = ?", hold.ID, "Active").
				Updates(map[string]interface{}{
					"status":      "Expired",
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
				count++
			}
		}
		return nil
	})
	return count, err
}

// StartHoldExpiryWorker launches a background goroutine that periodically expires overdue holds.
func StartHoldExpiryWorker(ctx context.Context, db *gorm.DB, interval time.Duration) {
	if interval <= 0 {
		interval = 1 * time.Minute
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				released, err := ExpireOverdueHolds(db)
				if err != nil {
					log.Printf("[HoldExpiryWorker] error expiring holds: %v\n", err)
				} else if released > 0 {
					log.Printf("[HoldExpiryWorker] expired %d overdue ticket hold(s)\n", released)
				}
			}
		}
	}()
}
