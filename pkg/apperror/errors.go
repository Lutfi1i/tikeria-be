package apperror

import (
	"errors"
	"net/http"
	"strings"
)

const (
	CodePackageLimitActiveEvent    = "PACKAGE_LIMIT_ACTIVE_EVENT"
	CodePackageLimitTicketCategory = "PACKAGE_LIMIT_TICKET_CATEGORY"
	CodePackageLimitQuota          = "PACKAGE_LIMIT_QUOTA"
	CodeEventLocked                = "EVENT_LOCKED"
	CodeHoldExpired                = "HOLD_EXPIRED"
	CodeTicketAlreadyUsed          = "TICKET_ALREADY_USED"
	CodeTicketNotFound             = "TICKET_NOT_FOUND"
	CodeWebViewTokenInvalid        = "WEBVIEW_TOKEN_INVALID"
	CodeForbidden                  = "FORBIDDEN"
	CodeNotFound                   = "NOT_FOUND"
)

type AppError struct {
	HTTPStatus int
	Code       string
	Message    string
}

func (e *AppError) Error() string {
	return e.Message
}

func New(httpStatus int, code, message string) *AppError {
	return &AppError{
		HTTPStatus: httpStatus,
		Code:       code,
		Message:    message,
	}
}

func NewPackageLimitActiveEvent() *AppError {
	return New(
		http.StatusForbidden,
		CodePackageLimitActiveEvent,
		"PACKAGE_LIMIT_ACTIVE_EVENT: Basic package allows 1 active event. Upgrade to Pro or wait until the current event ends.",
	)
}

func NewPackageLimitTicketCategory() *AppError {
	return New(
		http.StatusForbidden,
		CodePackageLimitTicketCategory,
		"PACKAGE_LIMIT_TICKET_CATEGORY: Basic package allows maximum 1 ticket category. Upgrade to Pro for unlimited categories.",
	)
}

func NewPackageLimitQuota() *AppError {
	return New(
		http.StatusForbidden,
		CodePackageLimitQuota,
		"PACKAGE_LIMIT_QUOTA: Basic package allows max 1.000 tickets per event",
	)
}

func NewEventLocked(message string) *AppError {
	if message == "" {
		message = "EVENT_LOCKED: event cannot be edited while Waiting Approval or Published"
	}
	return New(http.StatusUnprocessableEntity, CodeEventLocked, message)
}

func FromError(err error) (int, string, string) {
	if err == nil {
		return http.StatusOK, "", ""
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPStatus, appErr.Code, appErr.Message
	}

	msg := err.Error()
	switch {
	case strings.Contains(msg, CodePackageLimitActiveEvent):
		return http.StatusForbidden, CodePackageLimitActiveEvent, msg
	case strings.Contains(msg, CodePackageLimitTicketCategory):
		return http.StatusForbidden, CodePackageLimitTicketCategory, msg
	case strings.Contains(msg, CodePackageLimitQuota):
		return http.StatusForbidden, CodePackageLimitQuota, msg
	case strings.Contains(msg, CodeEventLocked):
		return http.StatusUnprocessableEntity, CodeEventLocked, msg
	case strings.Contains(msg, CodeHoldExpired):
		return http.StatusGone, CodeHoldExpired, msg
	case strings.Contains(msg, CodeWebViewTokenInvalid):
		return http.StatusUnauthorized, CodeWebViewTokenInvalid, msg
	case strings.Contains(msg, "unauthorized"):
		return http.StatusForbidden, CodeForbidden, msg
	case strings.Contains(msg, "not found"):
		return http.StatusNotFound, CodeNotFound, msg
	default:
		return http.StatusBadRequest, "", msg
	}
}
