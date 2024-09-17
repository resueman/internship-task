package service

import (
	"context"
	"tender-management-api/internal/entity"
	"tender-management-api/internal/repo"
)

type Diagnostics interface {
	Ping() error
}

type Tender interface {
	CreateTender(ctx context.Context, input *entity.CreateTenderInput) (*entity.TenderOutputModel, error)
	EditTenderById(ctx context.Context, tenderId string, username, name, description, serviceType string) (*entity.TenderOutputModel, error)

	GetTenderStatusById(ctx context.Context, tenderId string, username string, usernamePassed bool) (string, error)
	UpdateTenderStatusById(ctx context.Context, tenderId string, newStatus, username string) (*entity.TenderOutputModel, error)

	GetUserTenders(ctx context.Context, username string, usernamePassed bool, pg *entity.PaginationInput) ([]entity.TenderOutputModel, error)
	GetPublishedTenders(ctx context.Context, serviceTypes []string, pg *entity.PaginationInput) ([]entity.TenderOutputModel, error)

	RollbackTenderVersion(ctx context.Context, tenderId string, version int, username string) (*entity.TenderOutputModel, error)
}

type Bid interface {
	CreateBid(ctx context.Context, input *entity.CreateBidInput) (*entity.BidOutputModel, error)
	EditBidById(ctx context.Context, bidId string, username, name, description string) (*entity.BidOutputModel, error)

	GetBidStatusById(ctx context.Context, bidId string, username string) (string, error)
	UpdateBidStatusById(ctx context.Context, bidId string, newStatus, username string) (*entity.BidOutputModel, error)

	GetUserBids(ctx context.Context, username string, pg *entity.PaginationInput) ([]entity.BidOutputModel, error)
	GetBidsForTenderById(ctx context.Context, tenderId string, pg *entity.PaginationInput, username string) ([]entity.BidOutputModel, error)

	SubmitBidDecision(ctx context.Context, bidId string, decision, username string) (*entity.BidOutputModel, error)

	RollbackBidVersion(ctx context.Context, bidId string, version int, username string) (*entity.BidOutputModel, error)

	GetReviewsOnBidAuthorBids(ctx context.Context, tenderId string, authorUsername string, requesterUsername string, pg *entity.PaginationInput) ([]entity.ReviewOutputModel, error)

	SubmitBidFeedback(ctx context.Context, bidId string, username string, content string) (*entity.BidOutputModel, error)
}

type Services struct {
	Diagnostics Diagnostics
	Tender      Tender
	Bid         Bid
}

func NewServices(repos *repo.Repositories) *Services {
	return &Services{
		Tender:      NewTenderService(repos),
		Bid:         NewBidService(repos),
		Diagnostics: NewDiagnosticsService(repos),
	}
}
