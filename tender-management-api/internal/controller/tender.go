package controller

import (
	"net/http"
	"strconv"
	"strings"
	"tender-management-api/internal/common"
	"tender-management-api/internal/entity"
	"tender-management-api/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo"
)

type tenderRoutesHandler struct {
	tenderService service.Tender
	validate      *validator.Validate
}

func newTenderRoutesHandler(outer *echo.Group, services *service.Services, v *validator.Validate) *tenderRoutesHandler {
	h := &tenderRoutesHandler{tenderService: services.Tender, validate: v}

	outer.GET("/tenders", h.GetTenders)
	outer.POST("/tenders/new", h.PostTender)
	outer.GET("/tenders/my", h.GetUserTenders)
	outer.GET("/tenders/:tenderId/status", h.GetTenderStatus)
	outer.PUT("/tenders/:tenderId/status", h.UpdateTenderStatus)
	outer.PATCH("/tenders/:tenderId/edit", h.EditTender)
	outer.PUT("/tenders/:tenderId/rollback/:version", h.RollbackTenderVersion)

	return h
}

type getTenderInput struct {
	Limit        int32    `query:"limit" validate:"gte=0,lte=50"`
	Offset       int32    `query:"offset" validate:"gte=0"`
	ServiceTypes []string `query:"service_type" validate:"dive,oneof=Construction Delivery Manufacture"`
}

func newGetTenderInput() getTenderInput {
	return getTenderInput{Limit: defaultLimit, Offset: defaultOffset, ServiceTypes: make([]string, 0)}
}

// /tenders
func (h *tenderRoutesHandler) GetTenders(c echo.Context) error {
	var input = newGetTenderInput()
	if err := c.Bind(&input); err != nil {
		if e := c.JSON(http.StatusBadRequest, errorResponse{"Input data is not formed correctly"}); e != nil {
			return e
		}

		return err
	}

	if err := h.validate.Struct(input); err != nil {
		if e := c.JSON(http.StatusBadRequest, errorResponse{getAllErrorMessages(err)}); e != nil {
			return e
		}

		return err
	}

	pg := entity.NewPaginationInput(int(input.Limit), int(input.Offset))
	tenders, err := h.tenderService.GetPublishedTenders(c.Request().Context(), input.ServiceTypes, pg)
	if err != nil {
		if e := c.JSON(http.StatusBadRequest, errorResponse{err.Error()}); e != nil {
			return e
		}

		return err
	}
	if e := c.JSON(http.StatusOK, tenders); e != nil {
		return e
	}

	return nil
}

type postTenderInput struct {
	Name            string `json:"name" validate:"required,max=100"`
	Description     string `json:"description" validate:"required,max=500"`
	ServiceType     string `json:"serviceType" validate:"required,oneof=Construction Delivery Manufacture"`
	OrganizationId  string `json:"organizationId" validate:"required,max=100"`
	CreatorUsername string `json:"creatorUsername" validate:"required"`
}

// /tenders/new
func (h *tenderRoutesHandler) PostTender(c echo.Context) error {
	var input postTenderInput
	if err := c.Bind(&input); err != nil {
		if e := c.JSON(http.StatusBadRequest, errorResponse{"Input data is not formed correctly"}); e != nil {
			return e
		}

		return err
	}

	if err := h.validate.Struct(input); err != nil {
		if e := c.JSON(http.StatusBadRequest, errorResponse{getAllErrorMessages(err)}); e != nil {
			return e
		}

		return err
	}

	model := &entity.CreateTenderInput{
		Name: input.Name, Description: input.Description, ServiceType: input.ServiceType,
		OrganizationId: input.OrganizationId, CreatorUsername: input.CreatorUsername,
	}

	tender, err := h.tenderService.CreateTender(c.Request().Context(), model)
	if err == nil {
		if e := c.JSON(http.StatusOK, tender); e != nil {
			return e
		}

		return err
	}

	switch err {
	case service.ErrEmployeeNotFound:
		if e := c.JSON(http.StatusUnauthorized, errorResponse{"There is no employee with given username"}); e != nil {
			return e
		}
	case service.ErrOrganizationNotFound:
		if e := c.JSON(http.StatusBadRequest, errorResponse{"There is no organization with given id"}); e != nil {
			return e
		}
	case service.ErrUserIsNotOrganizationResponsible:
		if e := c.JSON(http.StatusForbidden, errorResponse{"You can't create tender from given organization, because you are not responsible for it"}); e != nil {
			return e
		}
	default:
		if e := c.JSON(http.StatusBadRequest, errorResponse{"Error"}); e != nil {
			return e
		}
	}

	return nil
}

type getUserTendersInput struct {
	Limit    int32  `query:"limit" validate:"gte=0,lte=50"`
	Offset   int32  `query:"offset" validate:"gte=0"`
	Username string `query:"username" validate:""`
}

func newGetUserTendersInput() getUserTendersInput {
	return getUserTendersInput{Limit: defaultLimit, Offset: defaultOffset, Username: defaultUsername}
}

// /tenders/my
func (h *tenderRoutesHandler) GetUserTenders(c echo.Context) error {
	var input = newGetUserTendersInput()
	if err := c.Bind(&input); err != nil {
		if e := c.JSON(http.StatusBadRequest, errorResponse{"Input data is not formed correctly"}); e != nil {
			return e
		}

		return err
	}

	if err := h.validate.Struct(input); err != nil {
		if e := c.JSON(http.StatusBadRequest, errorResponse{getAllErrorMessages(err)}); e != nil {
			return e
		}

		return err
	}

	pg := entity.NewPaginationInput(int(input.Limit), int(input.Offset))
	usernamePassed := input.Username != defaultUsername
	tenders, err := h.tenderService.GetUserTenders(c.Request().Context(), input.Username, usernamePassed, pg)
	if err == nil {
		if e := c.JSON(http.StatusOK, tenders); e != nil {
			return e
		}

		return nil
	}

	switch err {
	case service.ErrEmployeeNotFound:
		if e := c.JSON(http.StatusUnauthorized, errorResponse{"There is no employee with given username"}); e != nil {
			return e
		}
	default:
		if e := c.JSON(http.StatusBadRequest, errorResponse{err.Error()}); e != nil {
			return e
		}
	}

	return err
}

type getTenderStatusInput struct {
	TenderId string `path:"tenderId" validate:"required,max=100"`
	Username string `query:"username" validate:""`
}

// /tenders/:tenderId/status
func (h *tenderRoutesHandler) GetTenderStatus(c echo.Context) error {
	var input getTenderStatusInput
	if err := c.Bind(&input); err != nil {
		if e := c.JSON(http.StatusBadRequest, errorResponse{"Input data is not formed correctly"}); e != nil {
			return e
		}

		return err
	}

	input.TenderId = c.Param("tenderId")
	if err := h.validate.Struct(input); err != nil {
		if e := c.JSON(http.StatusBadRequest, errorResponse{getAllErrorMessages(err)}); e != nil {
			return e
		}

		return err
	}

	usernamePassed := input.Username != defaultUsername
	status, err := h.tenderService.GetTenderStatusById(c.Request().Context(), c.Param("tenderId"), input.Username, usernamePassed)
	if err == nil {
		if e := c.JSON(http.StatusOK, status); e != nil {
			return e
		}

		return nil
	}

	switch err {
	case service.ErrTenderNotFound:
		if e := c.JSON(http.StatusNotFound, errorResponse{"There is no tender with given id"}); e != nil {
			return e
		}
	case service.ErrUserHasNoAccessToTender:
		if e := c.JSON(http.StatusForbidden, errorResponse{"Only responsible for tender's organization can see tender's status"}); e != nil {
			return e
		}
	case service.ErrEmployeeNotFound:
		if e := c.JSON(http.StatusUnauthorized, errorResponse{"There is no employee with given username"}); e != nil {
			return e
		}
	case service.ErrUnauthorizedTryToAccessWithEmployeeRights:
		if e := c.JSON(http.StatusForbidden, errorResponse{"Try to pass username"}); e != nil {
			return e
		}
	default:
		if e := c.JSON(http.StatusBadRequest, errorResponse{"Error"}); e != nil {
			return e
		}
	}

	return err
}

type updateTenderStatusInput struct {
	Username string `query:"username" validate:"required"`
	Status   string `query:"status" validate:"required,oneof=Created Published Closed"`
	TenderId string `param:"tenderId" validate:"max=100"`
}

// /tenders/:tenderId/status
func (h *tenderRoutesHandler) UpdateTenderStatus(c echo.Context) error {
	var input updateTenderStatusInput
	if err := c.Bind(&input); err != nil {
		msg := err.Error()
		if !strings.Contains(msg, "Request body can't be empty") {
			if e := c.JSON(http.StatusBadRequest, errorResponse{"Input data is not formed correctly"}); e != nil {
				return e
			}

			return err
		}
	}

	input.TenderId, input.Status, input.Username = c.Param("tenderId"), c.QueryParam("status"), c.QueryParam("username")
	if err := h.validate.Struct(input); err != nil {
		if e := c.JSON(http.StatusBadRequest, errorResponse{getAllErrorMessages(err)}); e != nil {
			return e
		}

		return err
	}

	tender, err := h.tenderService.UpdateTenderStatusById(c.Request().Context(), input.TenderId, input.Status, input.Username)
	if err == nil {
		if e := c.JSON(http.StatusOK, tender); e != nil {
			return e
		}

		return nil
	}

	switch err {
	case service.ErrTenderNotFound:
		if e := c.JSON(http.StatusNotFound, errorResponse{"There is no tender with given id"}); e != nil {
			return e
		}
	case service.ErrEmployeeNotFound:
		if e := c.JSON(http.StatusUnauthorized, errorResponse{"There is no employee with given username"}); e != nil {
			return e
		}
	case service.ErrUserHasNoAccessToTender:
		if e := c.JSON(http.StatusForbidden, errorResponse{"You have no enough rights to update tender status"}); e != nil {
			return e
		}
	default:
		if e := c.JSON(http.StatusBadRequest, errorResponse{"Error"}); e != nil {
			return e
		}
	}

	return err
}

type editTenderInput struct {
	TenderId    string `param:"tenderId" validate:"required,max=100"`
	Username    string `query:"username" validate:"required"`
	Name        string `json:"name" validate:"max=100"`
	Description string `json:"description" validate:"max=500"`
	ServiceType string `json:"serviceType" validate:"oneof=Construction Delivery Manufacture"`
}

// /tenders/:tenderId/edit
func (h *tenderRoutesHandler) EditTender(c echo.Context) error {
	var input editTenderInput
	if err := c.Bind(&input); err != nil {
		if e := c.JSON(http.StatusBadRequest, errorResponse{"Input data is not formed correctly"}); e != nil {
			return e
		}

		return err
	}

	input.Username = c.QueryParam("username")
	input.TenderId = c.Param("tenderId")
	if err := h.validate.Struct(input); err != nil {
		m := getAllErrorMessages(err)
		if input.ServiceType != "" || input.ServiceType == common.Construction || input.ServiceType == common.Delivery || input.ServiceType == common.Manufacture {
			if e := c.JSON(http.StatusBadRequest, errorResponse{m}); e != nil {
				return e
			}

			return err
		}
	}

	tender, err := h.tenderService.EditTenderById(c.Request().Context(), input.TenderId, input.Username, input.Name, input.Description, input.ServiceType)
	if err == nil {
		if e := c.JSON(http.StatusOK, tender); e != nil {
			return e
		}

		return nil
	}

	switch err {
	case service.ErrTenderNotFound:
		if e := c.JSON(http.StatusNotFound, errorResponse{"There is no tender with given id"}); e != nil {
			return e
		}
	case service.ErrEmployeeNotFound:
		if e := c.JSON(http.StatusUnauthorized, errorResponse{"There is no employee with given username"}); e != nil {
			return e
		}
	case service.ErrUserHasNoAccessToTender:
		if e := c.JSON(http.StatusForbidden, errorResponse{"You have no enough rights to edit tender"}); e != nil {
			return e
		}
	default:
		if e := c.JSON(http.StatusBadRequest, errorResponse{"Error"}); e != nil {
			return e
		}
	}

	return err
}

type rollbackTenderVersionInput struct {
	TenderId string `param:"tenderId" validate:"required,max=100"`
	Version  int    `param:"version" validate:"required,min=1"`
	Username string `query:"username" validate:"required"`
}

// /tenders/:tenderId/rollback/:version
func (h *tenderRoutesHandler) RollbackTenderVersion(c echo.Context) error {
	var input rollbackTenderVersionInput
	if err := c.Bind(&input); err != nil {
		msg := err.Error()
		if !strings.Contains(msg, "Request body can't be empty") {
			if e := c.JSON(http.StatusBadRequest, errorResponse{"Input data is not formed correctly"}); e != nil {
				return e
			}

			return err
		}
	}

	v, _ := strconv.Atoi(c.Param("version"))
	input.TenderId, input.Username, input.Version = c.Param("tenderId"), c.QueryParam("username"), v
	if err := h.validate.Struct(input); err != nil {
		if e := c.JSON(http.StatusBadRequest, errorResponse{getAllErrorMessages(err)}); e != nil {
			return e
		}

		return err
	}

	tender, err := h.tenderService.RollbackTenderVersion(c.Request().Context(), input.TenderId, input.Version, input.Username)
	if err == nil {
		if e := c.JSON(http.StatusOK, tender); e != nil {
			return e
		}

		return nil
	}

	switch err {
	case service.ErrTenderNotFound:
		if e := c.JSON(http.StatusNotFound, errorResponse{"There is no tender with given id"}); e != nil {
			return e
		}
	case service.ErrEmployeeNotFound:
		if e := c.JSON(http.StatusUnauthorized, errorResponse{"There is no employee with given username"}); e != nil {
			return e
		}
	case service.ErrUserHasNoAccessToTender:
		if e := c.JSON(http.StatusForbidden, errorResponse{"You have no enough rights to rollaback / edit tender"}); e != nil {
			return e
		}
	case service.ErrNoSuchVersion:
		if e := c.JSON(http.StatusBadRequest, errorResponse{"No such version"}); e != nil {
			return e
		}
	default:
		if e := c.JSON(http.StatusBadRequest, errorResponse{"Error"}); e != nil {
			return e
		}
	}

	return err
}
