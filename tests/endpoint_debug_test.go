package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tikeria-be/controllers"
	"tikeria-be/dto"
	"tikeria-be/middleware"
	"tikeria-be/models"
	"tikeria-be/pkg/frappe"
	"tikeria-be/repositories"
	"tikeria-be/services"
	"tikeria-be/utils"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.Event{}, &models.Ticket{}); err != nil {
		t.Fatalf("failed to migrate models: %v", err)
	}

	userRepo := repositories.NewUserRepository(db)
	eventRepo := repositories.NewEventRepository(db)

	authService := services.NewAuthService(userRepo, eventRepo)
	eventService := services.NewEventService(eventRepo, userRepo)

	authController := controllers.NewAuthController(authService)
	eventController := controllers.NewEventController(eventService)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	mountGroup := func(group *gin.RouterGroup) {
		auth := group.Group("/auth")
		{
			auth.POST("/register", authController.Register)
			auth.POST("/login", authController.Login)
			auth.GET("/me", middleware.AuthMiddleware(), authController.GetMe)
		}
		organizer := group.Group("/organizer")
		{
			organizer.POST("/register", authController.RegisterOrganizer)
			organizer.POST("/verify-email", authController.VerifyEmail)
			organizer.GET("/me", middleware.AuthMiddleware("organizer", "partner", "admin"), authController.GetMe)
			organizer.GET("/package", middleware.AuthMiddleware("organizer", "partner", "admin"), authController.GetOrganizerPackage)
			organizer.GET("/events", middleware.AuthMiddleware("organizer", "partner", "admin"), eventController.GetOrganizerEvents)
			organizer.POST("/events", middleware.AuthMiddleware("organizer", "partner", "admin"), eventController.CreateEvent)
			organizer.PUT("/events/:id", middleware.AuthMiddleware("organizer", "partner", "admin"), eventController.UpdateEvent)
			organizer.POST("/events/:id/categories", middleware.AuthMiddleware("organizer", "partner", "admin"), eventController.AddTicketCategory)
			organizer.PUT("/events/:id/categories/:category_id", middleware.AuthMiddleware("organizer", "partner", "admin"), eventController.UpdateTicketCategory)
			organizer.DELETE("/events/:id/categories/:category_id", middleware.AuthMiddleware("organizer", "partner", "admin"), eventController.DeleteTicketCategory)
			organizer.POST("/events/:id/submit-approval", middleware.AuthMiddleware("organizer", "partner", "admin"), eventController.SubmitEventApproval)
		}
		admin := group.Group("/admin")
		{
			admin.POST("/events/:id/approve", middleware.AuthMiddleware("admin"), eventController.ApproveEvent)
			admin.POST("/events/:id/reject", middleware.AuthMiddleware("admin"), eventController.RejectEvent)
		}
		events := group.Group("/events")
		{
			events.GET("", eventController.GetAllEvents)
			events.GET("/:id", eventController.GetEventByID)
			events.GET("/:id/categories", eventController.GetEventCategories)
			events.POST("", middleware.AuthMiddleware("organizer", "partner", "admin"), eventController.CreateEvent)
		}
	}

	mountGroup(r.Group("/api"))
	mountGroup(r.Group("/api/v1"))

	return r, db
}

func doJSONRequest(r *gin.Engine, method, path string, token string, body interface{}) (*httptest.ResponseRecorder, dto.WebResponse) {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp dto.WebResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	return w, resp
}

func TestAllBackendEndpointsDebug(t *testing.T) {
	r, _ := setupTestRouter(t)

	// 1. Test CORS OPTIONS preflight returns 204 No Content
	wOpt, _ := doJSONRequest(r, http.MethodOptions, "/api/events", "", nil)
	if wOpt.Code != http.StatusNoContent {
		t.Fatalf("Expected OPTIONS status 204, got %d", wOpt.Code)
	}

	// 2. Test Organizer Registration: reject public_later in MVP
	wPub, respPub := doJSONRequest(r, http.MethodPost, "/api/organizer/register", "", map[string]interface{}{
		"organizer_name": "Public Promo",
		"email":          "pub@promo.id",
		"password":       "secret123",
		"whatsapp":       "081234567890",
		"organizer_type": "public_later",
		"institution":    "Umum",
		"city":           "Jakarta",
	})
	if wPub.Code != http.StatusBadRequest || respPub.Success {
		t.Fatalf("Expected 400 on public_later organizer_type, got %d (%+v)", wPub.Code, respPub)
	}

	// 3. Test Valid Organizer Registration (/api/organizer/register)
	wOrgReg, respOrgReg := doJSONRequest(r, http.MethodPost, "/api/organizer/register", "", map[string]interface{}{
		"organizer_name": "OSIS SMAN 1 Jakarta",
		"email":          "osis@sman1.sch.id",
		"password":       "secret123",
		"whatsapp":       "081234567890",
		"organizer_type": "school_campus",
		"institution":    "SMA Negeri 1 Jakarta",
		"city":           "Jakarta",
	})
	if wOrgReg.Code != http.StatusCreated || !respOrgReg.Success {
		t.Fatalf("Organizer register failed: %d %+v", wOrgReg.Code, respOrgReg)
	}

	// 4. Test Verify Email (/api/organizer/verify-email)
	wVer, respVer := doJSONRequest(r, http.MethodPost, "/api/organizer/verify-email", "", map[string]interface{}{
		"email": "osis@sman1.sch.id",
		"otp":   "123456",
	})
	if wVer.Code != http.StatusOK || !respVer.Success {
		t.Fatalf("Verify email failed: %d %+v", wVer.Code, respVer)
	}

	// 5. Login Organizer (/api/auth/login)
	wLogin, _ := doJSONRequest(r, http.MethodPost, "/api/auth/login", "", map[string]interface{}{
		"email":    "osis@sman1.sch.id",
		"password": "secret123",
	})
	if wLogin.Code != http.StatusOK {
		t.Fatalf("Organizer login failed: %d %s", wLogin.Code, wLogin.Body.String())
	}
	var loginPayload struct {
		Data dto.LoginResponse `json:"data"`
	}
	_ = json.Unmarshal(wLogin.Body.Bytes(), &loginPayload)
	orgToken := loginPayload.Data.Token
	if orgToken == "" {
		t.Fatalf("Expected non-empty JWT token")
	}

	// 6. Test GET /api/organizer/me and GET /api/organizer/package
	wMe, respMe := doJSONRequest(r, http.MethodGet, "/api/organizer/me", orgToken, nil)
	if wMe.Code != http.StatusOK || !respMe.Success {
		t.Fatalf("GET /api/organizer/me failed: %d %+v", wMe.Code, respMe)
	}

	wPkg, _ := doJSONRequest(r, http.MethodGet, "/api/organizer/package", orgToken, nil)
	if wPkg.Code != http.StatusOK {
		t.Fatalf("GET /api/organizer/package failed: %d", wPkg.Code)
	}
	var pkgPayload struct {
		Data dto.OrganizerPackageResponse `json:"data"`
	}
	_ = json.Unmarshal(wPkg.Body.Bytes(), &pkgPayload)
	if pkgPayload.Data.Package != "basic" || pkgPayload.Data.MaxQuotaPerEvent != 1000 || !pkgPayload.Data.Usage.CanCreateEvent {
		t.Fatalf("Unexpected package response before event creation: %+v", pkgPayload.Data)
	}

	// 7. Create Event as Draft (/api/organizer/events)
	eventDate := time.Now().Add(72 * time.Hour).Format(time.RFC3339)
	wEv, respEv := doJSONRequest(r, http.MethodPost, "/api/organizer/events", orgToken, map[string]interface{}{
		"title":          "Pensi SMAN 1 2026",
		"description":    "Pentas Seni Tahunan SMAN 1 Jakarta",
		"city":           "Jakarta",
		"venue_name":     "Lapangan Utama SMAN 1",
		"venue_address":  "Jl. Budi Utomo No. 7",
		"event_template": "pensi",
		"pic_name":       "Bapak Hendra",
		"pic_whatsapp":   "081199887766",
		"event_date":     eventDate,
	})
	if wEv.Code != http.StatusCreated || !respEv.Success {
		t.Fatalf("Create event failed: %d %+v", wEv.Code, respEv)
	}

	// 8. Verify Basic Package Active Event Limit: creating a 2nd active event must fail with 403 PACKAGE_LIMIT_ACTIVE_EVENT
	wEv2, respEv2 := doJSONRequest(r, http.MethodPost, "/api/organizer/events", orgToken, map[string]interface{}{
		"title":       "Pensi Kedua",
		"description": "Event kedua",
		"event_date":  eventDate,
	})
	if wEv2.Code != http.StatusForbidden || respEv2.ErrorCode != "PACKAGE_LIMIT_ACTIVE_EVENT" {
		t.Fatalf("Expected 403 PACKAGE_LIMIT_ACTIVE_EVENT on 2nd event, got %d (%+v)", wEv2.Code, respEv2)
	}

	// 9. Submit Approval before adding ticket category must fail with 400
	wSubEmpty, _ := doJSONRequest(r, http.MethodPost, "/api/organizer/events/1/submit-approval", orgToken, nil)
	if wSubEmpty.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 when submitting event with 0 categories, got %d", wSubEmpty.Code)
	}

	// 10. Add Ticket Category exceeding Basic Quota (1500 > 1000) must fail with 403 PACKAGE_LIMIT_QUOTA
	wCatQuota, respCatQuota := doJSONRequest(r, http.MethodPost, "/api/organizer/events/1/categories", orgToken, map[string]interface{}{
		"category_name": "Presale",
		"price":         50000,
		"quota":         1500,
	})
	if wCatQuota.Code != http.StatusForbidden || respCatQuota.ErrorCode != "PACKAGE_LIMIT_QUOTA" {
		t.Fatalf("Expected 403 PACKAGE_LIMIT_QUOTA, got %d (%+v)", wCatQuota.Code, respCatQuota)
	}

	// 11. Add Valid 1st Ticket Category (quota 500 <= 1000)
	wCat1, respCat1 := doJSONRequest(r, http.MethodPost, "/api/organizer/events/1/categories", orgToken, map[string]interface{}{
		"category_name": "Reguler Pensi",
		"price":         75000,
		"quota":         500,
	})
	if wCat1.Code != http.StatusCreated || !respCat1.Success {
		t.Fatalf("Expected 201 on 1st category, got %d (%+v)", wCat1.Code, respCat1)
	}

	// 12. Attempt to add 2nd Ticket Category on Basic Package -> 403 PACKAGE_LIMIT_TICKET_CATEGORY
	wCat2, respCat2 := doJSONRequest(r, http.MethodPost, "/api/organizer/events/1/categories", orgToken, map[string]interface{}{
		"category_name": "VIP Pensi",
		"price":         150000,
		"quota":         100,
	})
	if wCat2.Code != http.StatusForbidden || respCat2.ErrorCode != "PACKAGE_LIMIT_TICKET_CATEGORY" {
		t.Fatalf("Expected 403 PACKAGE_LIMIT_TICKET_CATEGORY on 2nd category, got %d (%+v)", wCat2.Code, respCat2)
	}

	// 13. Submit Event for Approval -> 200 OK (Status = Waiting Approval)
	wSub, respSub := doJSONRequest(r, http.MethodPost, "/api/organizer/events/1/submit-approval", orgToken, nil)
	if wSub.Code != http.StatusOK || !respSub.Success {
		t.Fatalf("Submit approval failed: %d (%+v)", wSub.Code, respSub)
	}

	// 14. Attempt to Edit Event while Waiting Approval -> 422 EVENT_LOCKED
	wEditLocked, respEditLocked := doJSONRequest(r, http.MethodPut, "/api/organizer/events/1", orgToken, map[string]interface{}{
		"title":       "Edited While Locked",
		"description": "Should fail",
	})
	if wEditLocked.Code != http.StatusUnprocessableEntity || respEditLocked.ErrorCode != "EVENT_LOCKED" {
		t.Fatalf("Expected 422 EVENT_LOCKED, got %d (%+v)", wEditLocked.Code, respEditLocked)
	}

	// 15. Admin Reject Event (requires reject_reason) & Organizer Resubmit -> Admin Approve -> Published
	adminToken, _ := utils.GenerateJWT(999, "admin@tikeria.id", "admin")

	wRejNoReason, _ := doJSONRequest(r, http.MethodPost, "/api/admin/events/1/reject", adminToken, map[string]interface{}{})
	if wRejNoReason.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 when rejecting without reason, got %d", wRejNoReason.Code)
	}

	wRej, respRej := doJSONRequest(r, http.MethodPost, "/api/admin/events/1/reject", adminToken, map[string]interface{}{
		"reject_reason": "Poster belum jelas, mohon perbaiki resolusi gambar",
	})
	if wRej.Code != http.StatusOK || !respRej.Success {
		t.Fatalf("Admin reject failed: %d (%+v)", wRej.Code, respRej)
	}

	// Organizer edits rejected event -> 200 OK
	wEditOk, respEditOk := doJSONRequest(r, http.MethodPut, "/api/organizer/events/1", orgToken, map[string]interface{}{
		"title":       "Pensi SMAN 1 2026 (Revisi HD)",
		"description": "Poster sudah diperbaiki",
		"banner_url":  "https://tikeria.id/poster-hd.jpg",
	})
	if wEditOk.Code != http.StatusOK || !respEditOk.Success {
		t.Fatalf("Expected 200 editing Rejected event, got %d (%+v)", wEditOk.Code, respEditOk)
	}

	// Organizer resubmits -> Waiting Approval -> Admin Approves -> Published
	wResub, _ := doJSONRequest(r, http.MethodPost, "/api/organizer/events/1/submit-approval", orgToken, nil)
	if wResub.Code != http.StatusOK {
		t.Fatalf("Resubmit failed: %d", wResub.Code)
	}

	wApp, respApp := doJSONRequest(r, http.MethodPost, "/api/admin/events/1/approve", adminToken, nil)
	if wApp.Code != http.StatusOK || !respApp.Success {
		t.Fatalf("Admin approve failed: %d (%+v)", wApp.Code, respApp)
	}

	// 16. Public GET /api/events, GET /api/events/1, GET /api/events/1/categories
	wPubList, _ := doJSONRequest(r, http.MethodGet, "/api/events", "", nil)
	if wPubList.Code != http.StatusOK {
		t.Fatalf("GET /api/events failed: %d", wPubList.Code)
	}

	wPubDet, _ := doJSONRequest(r, http.MethodGet, "/api/events/1", "", nil)
	if wPubDet.Code != http.StatusOK {
		t.Fatalf("GET /api/events/1 failed: %d", wPubDet.Code)
	}

	wPubCats, _ := doJSONRequest(r, http.MethodGet, "/api/events/1/categories", "", nil)
	if wPubCats.Code != http.StatusOK {
		t.Fatalf("GET /api/events/1/categories failed: %d", wPubCats.Code)
	}
}

func TestFrappeRESTHTTPClientEndToEnd(t *testing.T) {
	// Spin up an actual HTTP test server mimicking Frappe /api/resource/<DocType>
	mux := http.NewServeMux()
	mux.HandleFunc("/api/resource/", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "token test_key:test_secret" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"exc":"Unauthorized"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			body["name"] = "EVT-2026-00001"
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": body})
		case http.MethodGet:
			if r.URL.Query().Get("filters") != "" {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"data": []map[string]interface{}{
						{"name": "EVT-2026-00001", "title": "Pensi SMAN 1", "status": "Published"},
					},
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"name":   "EVT-2026-00001",
					"title":  "Pensi SMAN 1",
					"status": "Published",
				},
			})
		case http.MethodPut:
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			body["name"] = "EVT-2026-00001"
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": body})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := frappe.NewClient(srv.URL, "test_key", "test_secret")

	// 1. CreateDoc via HTTP POST
	created, err := client.CreateDoc("Tikeria Event", map[string]interface{}{
		"title":          "Pensi SMAN 1",
		"event_template": "pensi",
		"status":         "Draft",
	})
	if err != nil || fmt.Sprintf("%v", created["name"]) != "EVT-2026-00001" {
		t.Fatalf("CreateDoc HTTP failed: err=%v, doc=%+v", err, created)
	}

	// 2. GetDoc via HTTP GET
	fetched, err := client.GetDoc("Tikeria Event", "EVT-2026-00001")
	if err != nil || fmt.Sprintf("%v", fetched["status"]) != "Published" {
		t.Fatalf("GetDoc HTTP failed: err=%v, doc=%+v", err, fetched)
	}

	// 3. UpdateDoc via HTTP PUT
	updated, err := client.UpdateDoc("Tikeria Event", "EVT-2026-00001", map[string]interface{}{
		"status": "Waiting Approval",
	})
	if err != nil || fmt.Sprintf("%v", updated["status"]) != "Waiting Approval" {
		t.Fatalf("UpdateDoc HTTP failed: err=%v, doc=%+v", err, updated)
	}

	// 4. GetList via HTTP GET with filters & fields
	list, err := client.GetList("Tikeria Event", map[string]interface{}{"status": "Published"}, []string{"name", "title", "status"})
	if err != nil || len(list) != 1 {
		t.Fatalf("GetList HTTP failed: err=%v, list=%+v", err, list)
	}
}
