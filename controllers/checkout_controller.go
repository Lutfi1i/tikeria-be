package controllers

import (
	"net/http"
	"strings"

	"tikeria-be/dto"
	"tikeria-be/services"

	"github.com/gin-gonic/gin"
)

type CheckoutController struct {
	checkoutService services.CheckoutService
}

func NewCheckoutController(checkoutService services.CheckoutService) *CheckoutController {
	return &CheckoutController{checkoutService: checkoutService}
}

func getUserIDFromCtx(ctx *gin.Context) uint {
	if val, exists := ctx.Get("userID"); exists {
		if id, ok := val.(uint); ok {
			return id
		}
	}
	return 1 // Fallback for public/unprotected simulation in dev if any
}

// CreateHold godoc
// @Summary      Create Ticket Hold
// @Description  Create a 15-minute temporary ticket hold
// @Tags         Checkout
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateHoldRequest true "Hold Request"
// @Success      201  {object}  dto.WebResponse{data=dto.CreateHoldResponse}
// @Failure      400  {object}  dto.WebResponse
// @Failure      409  {object}  dto.WebResponse
// @Router       /api/checkout/hold [post]
func (c *CheckoutController) CreateHold(ctx *gin.Context) {
	var req dto.CreateHoldRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "Invalid request payload",
			Errors:  err.Error(),
		})
		return
	}

	userID := getUserIDFromCtx(ctx)
	resp, err := c.checkoutService.CreateHold(userID, req)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(strings.ToLower(err.Error()), "stock is not enough") ||
			strings.Contains(strings.ToLower(err.Error()), "insufficient") {
			status = http.StatusConflict
		}
		ctx.JSON(status, dto.WebResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, dto.WebResponse{
		Success: true,
		Message: "Ticket hold created successfully",
		Data:    resp,
	})
}

// ReleaseHold godoc
// @Summary      Release Ticket Hold
// @Description  Cancel an active hold and release ticket inventory
// @Tags         Checkout
// @Produce      json
// @Param        id   path      string  true  "Hold ID or Code"
// @Success      200  {object}  dto.WebResponse
// @Failure      400  {object}  dto.WebResponse
// @Router       /api/checkout/hold/{id} [delete]
func (c *CheckoutController) ReleaseHold(ctx *gin.Context) {
	holdID := ctx.Param("id")
	if strings.TrimSpace(holdID) == "" {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "hold id parameter is required",
		})
		return
	}

	userID := getUserIDFromCtx(ctx)
	if err := c.checkoutService.ReleaseHold(userID, holdID); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.WebResponse{
		Success: true,
		Message: "Ticket hold released successfully",
	})
}

// SaveParticipant godoc
// @Summary      Save Participant Detail
// @Description  Save attendee information for an active hold
// @Tags         Checkout
// @Accept       json
// @Produce      json
// @Param        request body dto.SaveParticipantRequest true "Participant details"
// @Success      200  {object}  dto.WebResponse{data=dto.SaveParticipantResponse}
// @Failure      400  {object}  dto.WebResponse
// @Router       /api/checkout/participant [post]
func (c *CheckoutController) SaveParticipant(ctx *gin.Context) {
	var req dto.SaveParticipantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "Invalid participant details payload",
			Errors:  err.Error(),
		})
		return
	}

	userID := getUserIDFromCtx(ctx)
	resp, err := c.checkoutService.SaveParticipant(userID, req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.WebResponse{
		Success: true,
		Message: "Participant detail saved successfully",
		Data:    resp,
	})
}

// GetSummary godoc
// @Summary      Checkout Summary
// @Description  Get order breakdown, flat platform service fee, and total amount
// @Tags         Checkout
// @Accept       json
// @Produce      json
// @Param        request body dto.CheckoutSummaryRequest true "Hold ID"
// @Success      200  {object}  dto.WebResponse{data=dto.CheckoutSummaryResponse}
// @Failure      400  {object}  dto.WebResponse
// @Router       /api/checkout/summary [post]
func (c *CheckoutController) GetSummary(ctx *gin.Context) {
	var req dto.CheckoutSummaryRequest
	holdID := ctx.Param("hold_id")
	if holdID == "" {
		if err := ctx.ShouldBindJSON(&req); err == nil && req.HoldID != nil {
			holdID = strings.TrimSpace(ctx.GetString("hold_id"))
			if holdID == "" {
				holdID = strings.TrimSpace(ctx.Query("hold_id"))
			}
			if holdID == "" {
				holdID = strings.TrimSpace(ctx.Param("id"))
			}
			if holdID == "" {
				holdID = strings.TrimSpace(string(ctx.GetHeader("X-Hold-ID")))
			}
			if holdID == "" {
				holdID = strings.TrimSpace(req.HoldID.(string))
			}
		}
	}

	if holdID == "" && req.HoldID != nil {
		holdID = strings.TrimSpace(req.HoldID.(string))
	}

	if holdID == "" {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "hold_id is required",
		})
		return
	}

	resp, err := c.checkoutService.GetSummary(holdID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.WebResponse{
		Success: true,
		Message: "Request success",
		Data:    resp,
	})
}

// CreateOrder godoc
// @Summary      Create Order
// @Description  Create order from active hold with participant details
// @Tags         Checkout
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateOrderRequest true "Order Request"
// @Success      201  {object}  dto.WebResponse{data=dto.CreateOrderResponse}
// @Failure      400  {object}  dto.WebResponse
// @Router       /api/checkout/create-order [post]
func (c *CheckoutController) CreateOrder(ctx *gin.Context) {
	var req dto.CreateOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "Invalid order creation payload",
			Errors:  err.Error(),
		})
		return
	}

	userID := getUserIDFromCtx(ctx)
	resp, err := c.checkoutService.CreateOrder(userID, req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, dto.WebResponse{
		Success: true,
		Message: "Order created successfully",
		Data:    resp,
	})
}

// GetOrderStatus godoc
// @Summary      Get Order Status
// @Description  Check order and payment status
// @Tags         Checkout
// @Produce      json
// @Param        id   path      string  true  "Order ID or Invoice Number"
// @Success      200  {object}  dto.WebResponse{data=dto.OrderStatusResponse}
// @Failure      404  {object}  dto.WebResponse
// @Router       /api/checkout/orders/{id}/status [get]
func (c *CheckoutController) GetOrderStatus(ctx *gin.Context) {
	identifier := ctx.Param("id")
	if identifier == "" {
		identifier = ctx.Param("order_id")
	}
	if identifier == "" {
		identifier = ctx.Param("invoice")
	}

	if strings.TrimSpace(identifier) == "" {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "order identifier or invoice number is required",
		})
		return
	}

	resp, err := c.checkoutService.GetOrderStatus(identifier)
	if err != nil {
		ctx.JSON(http.StatusNotFound, dto.WebResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.WebResponse{
		Success: true,
		Message: "Request success",
		Data:    resp,
	})
}

// SimulatePaymentSuccess godoc
// @Summary      Simulate Payment Success
// @Description  Mock gateway success webhook/callback
// @Tags         Payment
// @Produce      json
// @Param        order_id path string true "Order ID or Code"
// @Success      200 {object} dto.WebResponse{data=dto.OrderStatusResponse}
// @Router       /api/payments/{order_id}/simulate-success [post]
func (c *CheckoutController) SimulatePaymentSuccess(ctx *gin.Context) {
	orderID := ctx.Param("order_id")
	if orderID == "" {
		orderID = ctx.Param("id")
	}

	resp, err := c.checkoutService.SimulatePaymentSuccess(orderID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.WebResponse{
		Success: true,
		Message: "Payment simulation success",
		Data:    resp,
	})
}
