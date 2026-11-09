package frappe

import (
	"fmt"
	"log"

	"tikeria-be/config"
)

// SyncDocAsync sends a document to Frappe ERP asynchronously when FRAPPE_SYNC_ENABLED=true.
func SyncDocAsync(doctype string, docName string, payload map[string]interface{}) {
	cfg := config.LoadConfig()
	if !cfg.FrappeSyncEnabled {
		return
	}

	go func() {
		client := NewClient(cfg.FrappeBaseURL, cfg.FrappeAPIKey, cfg.FrappeAPISecret)
		payload["name"] = docName
		if _, err := client.CreateDoc(doctype, payload); err != nil {
			// Try UpdateDoc if already exists
			if _, updErr := client.UpdateDoc(doctype, docName, payload); updErr != nil {
				log.Printf("[Frappe Sync] warning syncing %s (%s): %v", doctype, docName, err)
			} else {
				log.Printf("[Frappe Sync] updated %s -> %s", doctype, docName)
			}
		} else {
			log.Printf("[Frappe Sync] synced %s -> %s", doctype, docName)
		}
	}()
}

//FormatEventDocName returns standard EVT-YYYY-00001 format.
func FormatEventDocName(year int, id uint) string {
	if year <= 0 {
		year = 2026
	}
	return fmt.Sprintf("EVT-%d-%05d", year, id)
}
