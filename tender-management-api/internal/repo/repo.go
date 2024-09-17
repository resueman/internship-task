package repo

import (
	"context"
	"tender-management-api/internal/entity"
	"tender-management-api/internal/repo/pgdb"
	"tender-management-api/pkg/postgres"

	"github.com/google/uuid"
)

type Diagnostics interface {
	Ping() error
}

type Employee interface {
	GetEmployeeIdByUsername(ctx context.Context, username string) (string, error)
	GetUserOrganizationIdByEmployeeId(ctx context.Context, employeeId string) (uuid.UUID, error)
	DoesOrganizationExistById(ctx context.Context, id string) (bool, error)
	DoesEmployeeExistsById(ctx context.Context, id string) (bool, error)
	IsEmployeeResponsible(ctx context.Context, employeeId string, organizationId uuid.UUID) (bool, error)
}

type Tender interface {
	CreateTender(ctx context.Context, input *entity.CreateTenderInput) (uuid.UUID, error)
	GetTenderById(ctx context.Context, id string) (*entity.Tender, error)
	EditTenderById(ctx context.Context, id string, name string, description string, serviceType string) error
	UpdateTenderStatusById(ctx context.Context, id string, newStatus string) error
	GetPublishedTenders(ctx context.Context, serviceTypes []string, pg *entity.PaginationInput) ([]entity.Tender, error)
	GetTendersByOrganizationId(ctx context.Context, organizationIds uuid.UUID, pg *entity.PaginationInput) ([]entity.Tender, error)
	RollbackTenderVersion(ctx context.Context, tenderId string, version int) error
}

type Bid interface {
	CreateBid(ctx context.Context, input *entity.CreateBidInput) (uuid.UUID, error)
	GetBidById(ctx context.Context, id string) (*entity.Bid, error)
	EditBidById(ctx context.Context, id string, name string, description string) error
	UpdateBidStatusById(ctx context.Context, id string, newStatus string) error
	GetUserBids(ctx context.Context, employeeId string, pg *entity.PaginationInput) ([]entity.Bid, error)
	GetTenderBids(ctx context.Context, tenderId string, pg *entity.PaginationInput) ([]entity.Bid, error)
	SubmitBidDecision(ctx context.Context, bidId string, decision string, employeeId string, organizationId uuid.UUID) error
	RollbackBidVersion(ctx context.Context, bidId string, version int) error
	SubmitBidFeedBack(ctx context.Context, bidId string, senderId uuid.UUID, receiverId uuid.UUID, content string) error
	GetReviewsByReceiverId(ctx context.Context, receiverId string, pg *entity.PaginationInput) ([]entity.Review, error)
	AlreadySubmitApprove(ctx context.Context, bidId string, employeeId string) (bool, error)
}

type Repositories struct {
	Diagnostics
	Employee
	Tender
	Bid
}

func NewRepositories(p *postgres.Postgres) *Repositories {
	return &Repositories{
		Diagnostics: pgdb.NewDiagnosticsRepo(p),
		Employee:    pgdb.NewEmployeeRepo(p),
		Tender:      pgdb.NewTenderRepo(p),
		Bid:         pgdb.NewBidRepo(p),
	}
}
