package controller

import (
	"net/http"
	"tender-management-api/internal/service"

	"github.com/labstack/echo"
)

type diagnosticRoutesHandler struct {
	diagnosticService service.Diagnostics
}

func newDiagnosticRoutesHandler(outer *echo.Group, services *service.Services) *diagnosticRoutesHandler {
	h := &diagnosticRoutesHandler{services.Diagnostics}
	outer.GET("/ping", h.Ping)

	return h
}

func (h *diagnosticRoutesHandler) Ping(c echo.Context) error {
	err := h.diagnosticService.Ping()
	if err != nil {
		if e := c.NoContent(http.StatusInternalServerError); e != nil {
			return e
		}

		return err
	}
	if e := c.JSON(http.StatusOK, "ok"); e != nil {
		return e
	}

	return nil
}
