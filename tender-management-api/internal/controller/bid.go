package controller

import (
	"net/http"
	"strconv"
	"strings"
	"tender-management-api/internal/entity"
	"tender-management-api/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo"
)

type bidRoutesHandler struct {
	bidService service.Bid
	validate   *validator.Validate
}

func newBidRoutesHandler(outer *echo.Group, services *service.Services, v *validator.Validate) *bidRoutesHandler {
	h := &bidRoutesHandler{bidService: services.Bid, validate: v}
	outer.POST("/bids/new", h.PostBid)
	outer.GET("/bids/my", h.GetUserBids)
	outer.GET("/bids/:tenderId/list", h.GetTenderBids)

	outer.GET("/bids/:bidId/status", h.GetBidStatus)
	outer.PUT("/bids/:bidId/status", h.UpdateBidStatus)

	outer.PATCH("/bids/:bidId/edit", h.EditBid)
	outer.PUT("/bids/:bidId/submit_decision", h.SubmitDecision)

	outer.PUT("/bids/:bidId/feedback", h.SubmitBidFeedback)
	outer.PUT("/bids/:bidId/rollback/:version", h.RollbackBidVersion)
	outer.GET("/bids/:tenderId/reviews", h.GetReviewsOnBidAuthorBids)

	return h
}

type postBidInput struct {
	Name        string `json:"name" validate:"required,max=100"`
	Description string `json:"description" validate:"required,max=500"`
	TenderId    string `json:"tenderId" validate:"required,max=100"`
	AuthorType  string `json:"authorType" validate:"required,oneof=Organization User"`
	AuthorId    string `json:"authorId" validate:"required,max=100"`
}

// в api не хватает bad request (например могут передать неверный тип пользователя)
// мне надо убрать все badRequests?
// /bids/new
func (h *bidRoutesHandler) PostBid(c echo.Context) error {
	var input postBidInput
	if err := c.Bind(&input); err != nil {
		if e := c.JSON(http.StatusBadRequest, errorResponse{"Input data is not formed correctly"}); e != nil {
			return e
		}

		return err
	}

	if err := h.validate.Struct(input); err != nil {
		if e := c.JSON(http.StatusBadRequest, errorResponse{"Not enough values passed or incorrect input value passed"}); e != nil {
			return e
		}

		return err
	}

	model := &entity.CreateBidInput{
		Name: input.Name, Description: input.Description, TenderId: input.TenderId,
		AuthorId: input.AuthorId, AuthorType: input.AuthorType,
	}

	bid, err := h.bidService.CreateBid(c.Request().Context(), model)
	if err == nil {
		if e := c.JSON(http.StatusOK, bid); e != nil {
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
		if e := c.JSON(http.StatusUnauthorized, errorResponse{"There is no employee with given id"}); e != nil {
			return e
		}
	case service.ErrUserHasNoAccessToTender:
		if e := c.JSON(http.StatusForbidden, errorResponse{"Tender isn't published, so you can't create bid"}); e != nil {
			return e
		}
	case service.ErrBidCanNotBeProposedBySameOrganization:
		if e := c.JSON(http.StatusForbidden, errorResponse{"Bid can't be proposed on behalf of the organization that owns the tender"}); e != nil {
			return e
		}
	default:
		if e := c.JSON(http.StatusBadRequest, errorResponse{"Error"}); e != nil {
			return e
		}
	}

	return err
}

type getUserBidsInput struct {
	Limit    int32  `query:"limit" validate:"gte=0,lte=50"`
	Offset   int32  `query:"offset" validate:"gte=0"`
	Username string `query:"username" validate:""`
}

func newGetUserBidsInput() getUserBidsInput {
	return getUserBidsInput{Limit: defaultLimit, Offset: defaultOffset, Username: defaultUsername}
}

// в api не хватает bad request (например могут передать limit=1000)
// мне надо убрать все badRequests?
// /bids/my
func (h *bidRoutesHandler) GetUserBids(c echo.Context) error {
	var input = newGetUserBidsInput()
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

	if input.Username == defaultUsername {
		if e := c.JSON(http.StatusUnauthorized, errorResponse{"Please provide your username"}); e != nil {
			return e
		}

		return nil
	}

	pg := entity.NewPaginationInput(int(input.Limit), int(input.Offset))
	bids, err := h.bidService.GetUserBids(c.Request().Context(), input.Username, pg)
	if err == nil {
		if e := c.JSON(http.StatusOK, bids); e != nil {
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
		if e := c.JSON(http.StatusBadRequest, errorResponse{"Error"}); e != nil {
			return e
		}
	}

	return err
}

type getTenderBidsInput struct {
	TenderId string `param:"tenderId" validate:"required,max=100"`
	Username string `query:"username" validate:"required"`
	Limit    int32  `query:"limit" validate:"gte=0,lte=50"`
	Offset   int32  `query:"offset" validate:"gte=0"`
}

func newGetTenderBidsInput() getTenderBidsInput {
	return getTenderBidsInput{
		Limit:  defaultLimit,
		Offset: defaultOffset,
	}
}

// /bids/:tenderId/list
func (h *bidRoutesHandler) GetTenderBids(c echo.Context) error {
	var input = newGetTenderBidsInput()
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

	pg := entity.NewPaginationInput(int(input.Limit), int(input.Offset))
	bids, err := h.bidService.GetBidsForTenderById(c.Request().Context(), input.TenderId, pg, input.Username)
	if err == nil {
		if e := c.JSON(http.StatusOK, bids); e != nil {
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
		if e := c.JSON(http.StatusForbidden, errorResponse{"Only responsible for tender's organization can see bids of requested tender"}); e != nil {
			return e
		}
	default:
		if e := c.JSON(http.StatusBadRequest, errorResponse{"Error"}); e != nil {
			return e
		}
	}

	return err
}

type getBidStatusInput struct {
	BidId    string `param:"bidId" validate:"required,max=100"`
	Username string `query:"username" validate:"required"`
}

// в api не хватает bad request (например могут передать bidId длиннее 100)
// С другой стороны все равно в бд нет бида с таким id
// убираю bad requests?
// /bids/:bidId/status
func (h *bidRoutesHandler) GetBidStatus(c echo.Context) error {
	var input getBidStatusInput
	if err := c.Bind(&input); err != nil {
		if e := c.JSON(http.StatusBadRequest, errorResponse{"Input data is not formed correctly"}); e != nil {
			return e
		}

		return err
	}

	input.BidId = c.Param("bidId")
	if err := h.validate.Struct(input); err != nil {
		if e := c.JSON(http.StatusBadRequest, errorResponse{getAllErrorMessages(err)}); e != nil {
			return e
		}

		return err
	}

	status, err := h.bidService.GetBidStatusById(c.Request().Context(), input.BidId, input.Username)
	if err == nil {
		if e := c.JSON(http.StatusOK, status); e != nil {
			return e
		}

		return nil
	}

	switch err {
	case service.ErrBidNotFound:
		if e := c.JSON(http.StatusNotFound, errorResponse{"There is no tender with given id"}); e != nil {
			return e
		}
	case service.ErrEmployeeNotFound:
		if e := c.JSON(http.StatusUnauthorized, errorResponse{"There is no employee with given username"}); e != nil {
			return e
		}
	case service.ErrUserHasNoAccessToBid:
		if e := c.JSON(http.StatusForbidden, errorResponse{"Only bid author and responsible for tender's organization can view bid status"}); e != nil {
			return e
		}
	default:
		if e := c.JSON(http.StatusBadRequest, errorResponse{"Error"}); e != nil {
			return e
		}
	}

	return err
}

type updateBidStatusInput struct {
	BidId    string `param:"bidId" validate:"required,max=100"`
	Username string `query:"username" validate:"required"`
	Status   string `query:"status" validate:"required,oneof=Created Published Canceled"`
}

// /bids/:bidId/status
func (h *bidRoutesHandler) UpdateBidStatus(c echo.Context) error {
	var input updateBidStatusInput
	if err := c.Bind(&input); err != nil {
		msg := err.Error()
		if !strings.Contains(msg, "Request body can't be empty") {
			if e := c.JSON(http.StatusBadRequest, errorResponse{"Input data is not formed correctly"}); e != nil {
				return e
			}

			return err
		}
	}
	input.BidId, input.Status, input.Username = c.Param("bidId"), c.QueryParam("status"), c.QueryParam("username")
	if err := h.validate.Struct(input); err != nil {
		if e := c.JSON(http.StatusBadRequest, errorResponse{getAllErrorMessages(err)}); e != nil {
			return e
		}

		return err
	}

	bid, err := h.bidService.UpdateBidStatusById(c.Request().Context(), input.BidId, input.Status, input.Username)
	if err == nil {
		if e := c.JSON(http.StatusOK, bid); e != nil {
			return e
		}

		return nil
	}

	switch err {
	case service.ErrBidNotFound:
		if e := c.JSON(http.StatusNotFound, errorResponse{"There is no bid with given id"}); e != nil {
			return e
		}
	case service.ErrEmployeeNotFound:
		if e := c.JSON(http.StatusUnauthorized, errorResponse{"There is no employee with given username"}); e != nil {
			return e
		}
	case service.ErrUserHasNoAccessToBid:
		if e := c.JSON(http.StatusForbidden, errorResponse{"You have not enough rights to update bid status"}); e != nil {
			return e
		}
	default:
		if e := c.JSON(http.StatusBadRequest, errorResponse{"Error"}); e != nil {
			return e
		}
	}

	return err
}

type editBidInput struct {
	BidId       string `param:"bidId" validate:"required,max=100"`
	Username    string `query:"username" validate:"required"`
	Name        string `json:"name" validate:"max=100"`
	Description string `json:"description" validate:"max=500"`
}

// /bids/:bidId/edit
func (h *bidRoutesHandler) EditBid(c echo.Context) error {
	var input editBidInput
	if err := c.Bind(&input); err != nil {
		if e := c.JSON(http.StatusBadRequest, errorResponse{"Input data is not formed correctly"}); e != nil {
			return e
		}

		return err
	}

	input.Username = c.QueryParam("username")
	input.BidId = c.Param("bidId")
	if input.Name == "" && input.Description == "" {
		if e := c.JSON(http.StatusBadRequest, errorResponse{"Bid updates required, set bid's name and/or description"}); e != nil {
			return e
		}

		return nil
	}

	bid, err := h.bidService.EditBidById(c.Request().Context(), input.BidId, input.Username, input.Name, input.Description)
	if err == nil {
		if e := c.JSON(http.StatusOK, bid); e != nil {
			return e
		}

		return nil
	}

	switch err {
	case service.ErrBidNotFound:
		if e := c.JSON(http.StatusNotFound, errorResponse{"There is no bid with given id"}); e != nil {
			return e
		}
	case service.ErrEmployeeNotFound:
		if e := c.JSON(http.StatusUnauthorized, errorResponse{"There is no employee with given username"}); e != nil {
			return e
		}
	case service.ErrUserHasNoAccessToBid:
		if e := c.JSON(http.StatusForbidden, errorResponse{"You have not enough rights to edit bid"}); e != nil {
			return e
		}
	default:
		if e := c.JSON(http.StatusBadRequest, errorResponse{"Error"}); e != nil {
			return e
		}
	}

	return err
}

type submitBidDecisionInput struct {
	BidId       string `param:"bidId" validate:"required"`
	Username    string `query:"username" validate:"required"`
	BisDecision string `query:"decision" validate:"required,oneof=Approved Rejected"`
}

// /bids/:bidId/submit_decision
func (h *bidRoutesHandler) SubmitDecision(c echo.Context) error {
	var input submitBidDecisionInput
	if err := c.Bind(&input); err != nil {
		msg := err.Error()
		if !strings.Contains(msg, "Request body can't be empty") {
			if e := c.JSON(http.StatusBadRequest, errorResponse{"Input data is not formed correctly"}); e != nil {
				return e
			}

			return err
		}
	}

	input.BidId, input.BisDecision, input.Username = c.Param("bidId"), c.QueryParam("decision"), c.QueryParam("username")
	if err := h.validate.Struct(input); err != nil {
		if e := c.JSON(http.StatusBadRequest, errorResponse{getAllErrorMessages(err)}); e != nil {
			return e
		}

		return err
	}

	bid, err := h.bidService.SubmitBidDecision(c.Request().Context(), input.BidId, input.BisDecision, input.Username)
	if err == nil {
		if e := c.JSON(http.StatusOK, bid); e != nil {
			return e
		}

		return nil
	}

	switch err {
	case service.ErrBidNotFound:
		if e := c.JSON(http.StatusNotFound, errorResponse{"There is no bid with given id"}); e != nil {
			return e
		}
	case service.ErrEmployeeNotFound:
		if e := c.JSON(http.StatusUnauthorized, errorResponse{"There is no employee with given username"}); e != nil {
			return e
		}
	case service.ErrTenderNotFound:
		if e := c.JSON(http.StatusNotFound, errorResponse{"There is no more tender for bid"}); e != nil {
			return e
		}
	case service.ErrBidAuthorCanNotMakeDecisionsOnIt:
		if e := c.JSON(http.StatusForbidden, errorResponse{"You can't make decision on bid, because you are its author"}); e != nil {
			return e
		}
	case service.ErrUserHasNoAccessToTender:
		if e := c.JSON(http.StatusForbidden, errorResponse{"You aren't responsible for organization that opened tender, therefore you can't submit decisions connected with given tender"}); e != nil {
			return e
		}
	case service.ErrAlreadyApproveBid:
		if e := c.JSON(http.StatusForbidden, errorResponse{"You have already approved bid"}); e != nil {
			return e
		}
	default:
		if e := c.JSON(http.StatusBadRequest, errorResponse{"Error"}); e != nil {
			return e
		}
	}

	return err
}

type rollbackBidVersionInput struct {
	BidId    string `param:"bidId" validate:"required,max=100"`
	Version  int    `param:"version" validate:"required,min=1"`
	Username string `query:"username" validate:"required"`
}

// /bids/:bidId/rollback/:version
func (h *bidRoutesHandler) RollbackBidVersion(c echo.Context) error {
	var input rollbackBidVersionInput
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
	input.BidId, input.Username, input.Version = c.Param("bidId"), c.QueryParam("username"), v
	if err := h.validate.Struct(input); err != nil {
		if e := c.JSON(http.StatusBadRequest, errorResponse{getAllErrorMessages(err)}); e != nil {
			return e
		}

		return err
	}

	tender, err := h.bidService.RollbackBidVersion(c.Request().Context(), input.BidId, input.Version, input.Username)
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
	case service.ErrUserHasNoAccessToBid:
		if e := c.JSON(http.StatusForbidden, errorResponse{"You have no enough rights to edit tender"}); e != nil {
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

type getReviewsOnBidAuthorBidsInput struct {
	TenderId          string `param:"tenderId" validate:"required,max=100"`
	AuthorUsername    string `query:"authorUsername" validate:"required"`
	RequesterUsername string `query:"requesterUsername" validate:"required"`
	Limit             int32  `query:"limit" validate:"gte=0,lte=50"`
	Offset            int32  `query:"offset" validate:"gte=0"`
}

func newGetReviewsOnBidAuthorBidsInput() getReviewsOnBidAuthorBidsInput {
	return getReviewsOnBidAuthorBidsInput{
		Limit:  defaultLimit,
		Offset: defaultOffset,
	}
}

// /bids/:tenderId/reviews
func (h *bidRoutesHandler) GetReviewsOnBidAuthorBids(c echo.Context) error {
	var input = newGetReviewsOnBidAuthorBidsInput()
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

	pg := entity.NewPaginationInput(int(input.Limit), int(input.Offset))
	reviews, err := h.bidService.GetReviewsOnBidAuthorBids(c.Request().Context(),
		input.TenderId, input.AuthorUsername, input.RequesterUsername, pg)
	if err == nil {
		if e := c.JSON(http.StatusOK, reviews); e != nil {
			return e
		}

		return nil
	}

	switch err {
	case service.ErrBidAuthorNotAnEmployee:
		if e := c.JSON(http.StatusNotFound, errorResponse{"There is no employee with given username for author of bid"}); e != nil {
			return e
		}
	case service.ErrRequesterNotAnEmployee:
		if e := c.JSON(http.StatusUnauthorized, errorResponse{"There is no employee with given username for review requester"}); e != nil {
			return e
		}
	case service.ErrTenderNotFound:
		if e := c.JSON(http.StatusForbidden, errorResponse{"There is no tender with given id"}); e != nil {
			return e
		}
	case service.ErrUserHasNoAccessToTender:
		if e := c.JSON(http.StatusForbidden, errorResponse{"You have no enough rights to access tender => tender bid => reviews on bid author"}); e != nil {
			return e
		}
	default:
		if e := c.JSON(http.StatusBadRequest, errorResponse{"Error"}); e != nil {
			return e
		}
	}

	return err
}

type submitBidFeedbackInput struct {
	BidId    string `param:"bidId" validate:"required,max=100"`
	Username string `query:"username" validate:"required"`
	FeedBack string `query:"bidFeedback" validate:"required,max=1000"`
}

// /bids/:bidId/feedback
func (h *bidRoutesHandler) SubmitBidFeedback(c echo.Context) error {
	var input submitBidFeedbackInput
	if err := c.Bind(&input); err != nil {
		msg := err.Error()
		if !strings.Contains(msg, "Request body can't be empty") {
			if e := c.JSON(http.StatusBadRequest, errorResponse{"Input data is not formed correctly"}); e != nil {
				return e
			}

			return err
		}
	}

	input.BidId, input.FeedBack, input.Username = c.Param("bidId"), c.QueryParam("bidFeedback"), c.QueryParam("username")
	if err := h.validate.Struct(input); err != nil {
		if e := c.JSON(http.StatusBadRequest, errorResponse{getAllErrorMessages(err)}); e != nil {
			return e
		}

		return err
	}

	tender, err := h.bidService.SubmitBidFeedback(c.Request().Context(), input.BidId, input.Username, input.FeedBack)
	if err == nil {
		if e := c.JSON(http.StatusOK, tender); e != nil {
			return e
		}

		return nil
	}

	switch err {
	case service.ErrBidNotFound:
		if e := c.JSON(http.StatusNotFound, errorResponse{"There is no bid with given id"}); e != nil {
			return e
		}
	case service.ErrEmployeeNotFound:
		if e := c.JSON(http.StatusUnauthorized, errorResponse{"There is no employee with given username"}); e != nil {
			return e
		}
	case service.ErrUserHasNoAccessToBid:
		if e := c.JSON(http.StatusForbidden, errorResponse{"You have no enough rights to sumbit feedback to bid"}); e != nil {
			return e
		}
	default:
		if e := c.JSON(http.StatusBadRequest, errorResponse{"Error"}); e != nil {
			return e
		}
	}

	return err
}
