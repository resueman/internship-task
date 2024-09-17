package controller

import (
	"tender-management-api/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo"
)

func SetupRoutesHandlers(handler *echo.Echo, services *service.Services) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	api := handler.Group("/api")
	newDiagnosticRoutesHandler(api, services)
	newBidRoutesHandler(api, services, validate)
	newTenderRoutesHandler(api, services, validate)
}
