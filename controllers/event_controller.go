package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"tikeria-be/dto"
	"tikeria-be/services"
)

type EventController struct {
	eventService services.EventService
}

func NewEventController(eventService services.EventService) *EventController {
	return &EventController{eventService: eventService}
}

func mapServiceError(err error) (int, string) {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "PACKAGE_LIMIT_ACTIVE_EVENT"):
		return http.StatusForbidden, "PACKAGE_LIMIT_ACTIVE_EVENT"
	case strings.Contains(msg, "PACKAGE_LIMIT_TICKET_CATEGORY"):
		return http.StatusForbidden, "PACKAGE_LIMIT_TICKET_CATEGORY"
	case strings.Contains(msg, "PACKAGE_LIMIT_QUOTA"):
		return http.StatusForbidden, "PACKAGE_LIMIT_QUOTA"
	case strings.Contains(msg, "EVENT_LOCKED"):
		return http.StatusUnprocessableEntity, "EVENT_LOCKED"
	case strings.Contains(msg, "unauthorized"):
		return http.StatusForbidden, "FORBIDDEN"
	case strings.Contains(msg, "not found"):
		return http.StatusNotFound, "NOT_FOUND"
	default:
		return http.StatusBadRequest, ""
	}
}

// CreateEvent godoc
// @Summary      Make new Event
// @Description  Make a new Event with the option (Organizer / Partner / Admin)
// @Tags         Events
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateEventRequest true "Event & Ticket Data"
// @Success      201  {object}  dto.WebResponse{data=models.Event}
// @Failure      400  {object}  dto.WebResponse
// @Failure      401  {object}  dto.WebResponse
// @Failure      403  {object}  dto.WebResponse
// @Router       /api/v1/events [post]
func (c *EventController) CreateEvent(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.WebResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	var req dto.CreateEventRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "Format event data is not valid",
			Errors:  err.Error(),
		})
		return
	}

	// If called via /events public-compat route and no status is specified, keep compatibility;
	// if called via /organizer/events, default status is Draft.
	if req.Status == "" {
		if strings.Contains(ctx.FullPath(), "/organizer/events") {
			req.Status = "Draft"
		} else {
			req.Status = "Draft"
		}
	}

	isPro := strings.EqualFold(ctx.GetString("organizerPackage"), "pro")
	event, err := c.eventService.CreateEvent(userID.(uint), req, isPro)
	if err != nil {
		status, errCode := mapServiceError(err)
		ctx.JSON(status, dto.WebResponse{
			Success:   false,
			Message:   err.Error(),
			ErrorCode: errCode,
		})
		return
	}

	ctx.JSON(http.StatusCreated, dto.WebResponse{
		Success: true,
		Message: "Event Created Successfully",
		Data:    event,
	})
}

// UpdateEvent godoc
// @Summary      Update Event
// @Description  Update an existing Draft or Rejected Event owned by the organizer
// @Tags         Organizer, Events
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Event ID"
// @Param        request body dto.UpdateEventRequest true "Event Update Data"
// @Success      200  {object}  dto.WebResponse{data=models.Event}
// @Router       /api/v1/organizer/events/{id} [put]
func (c *EventController) UpdateEvent(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.WebResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	eventID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "Invalid Event ID",
		})
		return
	}

	var req dto.UpdateEventRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "Invalid event update input",
			Errors:  err.Error(),
		})
		return
	}

	event, err := c.eventService.UpdateEvent(userID.(uint), uint(eventID), req)
	if err != nil {
		status, errCode := mapServiceError(err)
		ctx.JSON(status, dto.WebResponse{
			Success:   false,
			Message:   err.Error(),
			ErrorCode: errCode,
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.WebResponse{
		Success: true,
		Message: "Event updated successfully",
		Data:    event,
	})
}

// GetAllEvents godoc
// @Summary      See All Event
// @Description  Get all event from list and ticket availability
// @Tags         Events
// @Produce      json
// @Success      200  {object}  dto.WebResponse{data=[]models.Event}
// @Router       /api/v1/events [get]
func (c *EventController) GetAllEvents(ctx *gin.Context) {
	events, err := c.eventService.GetAllEvents()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.WebResponse{
			Success: false,
			Message: "failed to get event",
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.WebResponse{
		Success: true,
		Message: "Success get all event",
		Data:    events,
	})
}

// GetEventByID godoc
// @Summary      Event Detail
// @Description  Get Event Detail From ID
// @Tags         Events
// @Produce      json
// @Param        id path int true "Event ID"
// @Success      200  {object}  dto.WebResponse{data=models.Event}
// @Failure      404  {object}  dto.WebResponse
// @Router       /api/v1/events/{id} [get]
func (c *EventController) GetEventByID(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "ID Event is not valid",
		})
		return
	}

	event, err := c.eventService.GetEventByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, dto.WebResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.WebResponse{
		Success: true,
		Message: "Success, fetch all Event",
		Data:    event,
	})
}

// GetEventCategories godoc
// @Summary      Event Categories
// @Description  Get Ticket Categories for an Event
// @Tags         Events
// @Produce      json
// @Param        id path int true "Event ID"
// @Success      200  {object}  dto.WebResponse{data=[]models.Ticket}
// @Router       /api/v1/events/{id}/categories [get]
func (c *EventController) GetEventCategories(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "Invalid Event ID",
		})
		return
	}

	event, err := c.eventService.GetEventByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, dto.WebResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.WebResponse{
		Success: true,
		Message: "Success get event categories",
		Data:    event.Tickets,
	})
}

// GetOrganizerEvents godoc
// @Summary      Get Organizer Events
// @Description  Get list of events created by authenticated organizer
// @Tags         Organizer, Events
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.WebResponse{data=[]models.Event}
// @Router       /api/v1/organizer/events [get]
func (c *EventController) GetOrganizerEvents(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.WebResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	events, err := c.eventService.GetEventsByOrganizer(userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.WebResponse{
			Success: false,
			Message: "Failed to retrieve organizer events",
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.WebResponse{
		Success: true,
		Message: "Success retrieved organizer events",
		Data:    events,
	})
}

// AddTicketCategory godoc
// @Summary      Add Ticket Category
// @Description  Add ticket category to organizer event with package limit enforcement
// @Tags         Organizer, Events
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Event ID"
// @Param        request body dto.TicketInput true "Ticket Category Data"
// @Success      201  {object}  dto.WebResponse{data=models.Ticket}
// @Failure      400  {object}  dto.WebResponse
// @Failure      403  {object}  dto.WebResponse
// @Router       /api/v1/organizer/events/{id}/categories [post]
func (c *EventController) AddTicketCategory(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.WebResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	idParam := ctx.Param("id")
	eventID, err := strconv.Atoi(idParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "Invalid Event ID",
		})
		return
	}

	var req dto.TicketInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "Invalid ticket category input",
			Errors:  err.Error(),
		})
		return
	}

	pkg := ctx.GetString("organizerPackage")
	isPro := strings.EqualFold(pkg, "pro")

	ticket, err := c.eventService.AddTicketCategory(userID.(uint), uint(eventID), req, isPro)
	if err != nil {
		status, errCode := mapServiceError(err)
		ctx.JSON(status, dto.WebResponse{
			Success:   false,
			Message:   err.Error(),
			ErrorCode: errCode,
		})
		return
	}

	ctx.JSON(http.StatusCreated, dto.WebResponse{
		Success: true,
		Message: "Ticket category added successfully",
		Data:    ticket,
	})
}

// UpdateTicketCategory godoc
// @Summary      Update Ticket Category
// @Description  Update ticket category of an organizer event
// @Tags         Organizer, Events
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Event ID"
// @Param        category_id path int true "Category ID"
// @Param        request body dto.TicketInput true "Ticket Category Data"
// @Success      200  {object}  dto.WebResponse{data=models.Ticket}
// @Router       /api/v1/organizer/events/{id}/categories/{category_id} [put]
func (c *EventController) UpdateTicketCategory(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.WebResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	eventID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "Invalid Event ID",
		})
		return
	}

	categoryID, err := strconv.Atoi(ctx.Param("category_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "Invalid Category ID",
		})
		return
	}

	var req dto.TicketInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "Invalid ticket category input",
			Errors:  err.Error(),
		})
		return
	}

	isPro := strings.EqualFold(ctx.GetString("organizerPackage"), "pro")
	ticket, err := c.eventService.UpdateTicketCategory(userID.(uint), uint(eventID), uint(categoryID), req, isPro)
	if err != nil {
		status, errCode := mapServiceError(err)
		ctx.JSON(status, dto.WebResponse{
			Success:   false,
			Message:   err.Error(),
			ErrorCode: errCode,
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.WebResponse{
		Success: true,
		Message: "Ticket category updated successfully",
		Data:    ticket,
	})
}

// DeleteTicketCategory godoc
// @Summary      Delete Ticket Category
// @Description  Delete ticket category of an organizer event
// @Tags         Organizer, Events
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Event ID"
// @Param        category_id path int true "Category ID"
// @Success      200  {object}  dto.WebResponse
// @Router       /api/v1/organizer/events/{id}/categories/{category_id} [delete]
func (c *EventController) DeleteTicketCategory(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.WebResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	eventID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "Invalid Event ID",
		})
		return
	}

	categoryID, err := strconv.Atoi(ctx.Param("category_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "Invalid Category ID",
		})
		return
	}

	if err := c.eventService.DeleteTicketCategory(userID.(uint), uint(eventID), uint(categoryID)); err != nil {
		status, errCode := mapServiceError(err)
		ctx.JSON(status, dto.WebResponse{
			Success:   false,
			Message:   err.Error(),
			ErrorCode: errCode,
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.WebResponse{
		Success: true,
		Message: "Ticket category deleted successfully",
	})
}

// SubmitEventApproval godoc
// @Summary      Submit Event for Approval
// @Description  Transition event from Draft/Reject to Waiting Approval
// @Tags         Organizer, Events
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Event ID"
// @Success      200  {object}  dto.WebResponse{data=models.Event}
// @Router       /api/v1/organizer/events/{id}/submit-approval [post]
func (c *EventController) SubmitEventApproval(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.WebResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	eventID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "Invalid Event ID",
		})
		return
	}

	event, err := c.eventService.SubmitEventApproval(userID.(uint), uint(eventID))
	if err != nil {
		status, errCode := mapServiceError(err)
		ctx.JSON(status, dto.WebResponse{
			Success:   false,
			Message:   err.Error(),
			ErrorCode: errCode,
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.WebResponse{
		Success: true,
		Message: "Event submitted for approval",
		Data:    event,
	})
}

// ApproveEvent godoc
// @Summary      Approve Event (Admin)
// @Description  Transition event from Waiting Approval to Published
// @Tags         Admin, Events
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Event ID"
// @Success      200  {object}  dto.WebResponse{data=models.Event}
// @Router       /api/v1/admin/events/{id}/approve [post]
func (c *EventController) ApproveEvent(ctx *gin.Context) {
	email := ctx.GetString("email")
	if email == "" {
		email = "admin@tikeria.id"
	}

	eventID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "Invalid Event ID",
		})
		return
	}

	event, err := c.eventService.ApproveEvent(email, uint(eventID))
	if err != nil {
		status, errCode := mapServiceError(err)
		ctx.JSON(status, dto.WebResponse{
			Success:   false,
			Message:   err.Error(),
			ErrorCode: errCode,
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.WebResponse{
		Success: true,
		Message: "Event approved and published",
		Data:    event,
	})
}

// RejectEvent godoc
// @Summary      Reject Event (Admin)
// @Description  Transition event from Waiting Approval to Reject with mandatory reject_reason
// @Tags         Admin, Events
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Event ID"
// @Param        request body dto.RejectEventRequest true "Reject Reason"
// @Success      200  {object}  dto.WebResponse{data=models.Event}
// @Router       /api/v1/admin/events/{id}/reject [post]
func (c *EventController) RejectEvent(ctx *gin.Context) {
	email := ctx.GetString("email")
	if email == "" {
		email = "admin@tikeria.id"
	}

	eventID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "Invalid Event ID",
		})
		return
	}

	var req dto.RejectEventRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "reject_reason is required when rejecting an event",
			Errors:  err.Error(),
		})
		return
	}

	event, err := c.eventService.RejectEvent(email, uint(eventID), req.RejectReason)
	if err != nil {
		status, errCode := mapServiceError(err)
		ctx.JSON(status, dto.WebResponse{
			Success:   false,
			Message:   err.Error(),
			ErrorCode: errCode,
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.WebResponse{
		Success: true,
		Message: "Event rejected",
		Data:    event,
	})
}
