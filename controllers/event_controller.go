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

// CreateEvent godoc
// @Summary      Make new Event
// @Description  Make a new Event with the option (Only Partner / Admin)
// @Tags         Events
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateEventRequest true "Event & Ticket Data"
// @Success      201  {object}  dto.WebResponse{data=models.Event}
// @Failure      400  {object}  dto.WebResponse
// @Failure      401  {object}  dto.WebResponse
// @Router       /api/v1/events [post]
func (c *EventController) CreateEvent(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")

	var req dto.CreateEventRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "Format event data is not valid",
			Errors:  err.Error(),
		})
		return
	}

	event, err := c.eventService.CreateEvent(userID.(uint), req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, dto.WebResponse{
		Success: true,
		Message: "Event Created Successfully",
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
	isPro := pkg == "pro"

	ticket, err := c.eventService.AddTicketCategory(userID.(uint), uint(eventID), req, isPro)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "unauthorized: you do not own this event" {
			status = http.StatusForbidden
		} else if strings.Contains(err.Error(), "package limit reached") {
			status = http.StatusForbidden
		}
		ctx.JSON(status, dto.WebResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, dto.WebResponse{
		Success: true,
		Message: "Ticket category added successfully",
		Data:    ticket,
	})
}