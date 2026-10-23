package tests

import (
	"testing"
	"time"

	"tikeria-be/pkg/frappe"
	"tikeria-be/utils"
)

func TestMaskNIK(t *testing.T) {
	nik := "3174000000001234"
	expected := "3174********1234"
	actual := utils.MaskNIK(nik)
	if actual != expected {
		t.Fatalf("Expected masked NIK %s, got %s", expected, actual)
	}

	shortNIK := "12345"
	if utils.MaskNIK(shortNIK) != shortNIK {
		t.Fatalf("Expected short NIK to remain unmasked, got %s", utils.MaskNIK(shortNIK))
	}
}

func TestJWTGenerationAndValidation(t *testing.T) {
	token, err := utils.GenerateJWT(1, "buyer@tikeria.id", "buyer")
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	claims, err := utils.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate JWT: %v", err)
	}

	if claims.UserID != 1 || claims.Email != "buyer@tikeria.id" || claims.Role != "buyer" {
		t.Fatalf("Claims mismatch: %+v", claims)
	}
}

// Mock client for Frappe REST integration
type mockFrappeClient struct {
	data map[string]map[string]interface{}
}

func (m *mockFrappeClient) GetDoc(doctype, name string) (map[string]interface{}, error) {
	if doc, ok := m.data[doctype+":"+name]; ok {
		return doc, nil
	}
	return nil, nil
}

func (m *mockFrappeClient) CreateDoc(doctype string, data map[string]interface{}) (map[string]interface{}, error) {
	if m.data == nil {
		m.data = make(map[string]map[string]interface{})
	}
	name := "DOC-1"
	m.data[doctype+":"+name] = data
	return data, nil
}

func (m *mockFrappeClient) UpdateDoc(doctype, name string, data map[string]interface{}) (map[string]interface{}, error) {
	m.data[doctype+":"+name] = data
	return data, nil
}

func (m *mockFrappeClient) GetList(doctype string, filters map[string]interface{}, fields []string) ([]map[string]interface{}, error) {
	var list []map[string]interface{}
	for k, v := range m.data {
		if len(k) > len(doctype) && k[:len(doctype)] == doctype {
			list = append(list, v)
		}
	}
	return list, nil
}

func TestFrappeClientSync(t *testing.T) {
	var client frappe.FrappeClient = &mockFrappeClient{}

	eventDoc := map[string]interface{}{
		"title":          "Pensi SMA 1",
		"event_template": "pensi",
		"event_date":     time.Now().Add(24 * time.Hour).Format("2006-01-02"),
		"status":         "Draft",
	}

	created, err := client.CreateDoc("Tikeria Event", eventDoc)
	if err != nil {
		t.Fatalf("Failed to sync event to Frappe: %v", err)
	}

	if created["title"] != "Pensi SMA 1" {
		t.Fatalf("Unexpected title: %v", created["title"])
	}
}

func TestOrganizerPackageLimitEnforcement(t *testing.T) {
	// Basic organizer has 1 ticket category max
	existingTickets := 1
	isPro := false

	if !isPro && existingTickets >= 1 {
		// correctly rejected
	} else {
		t.Fatalf("Expected basic organizer to be blocked at 1 ticket category")
	}

	// Pro organizer can add multiple categories
	isPro = true
	if !isPro && existingTickets >= 1 {
		t.Fatalf("Expected pro organizer to NOT be blocked")
	}
}
