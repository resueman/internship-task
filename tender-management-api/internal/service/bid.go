package service

import (
	"context"
	"errors"
	"tender-management-api/internal/common"
	"tender-management-api/internal/entity"
	"tender-management-api/internal/repo"
	"tender-management-api/internal/repo/repo_errors"

	"github.com/google/uuid"
)

type BidService struct {
	bidRepo      repo.Bid
	employeeRepo repo.Employee
	tenderRepo   repo.Tender
}

func NewBidService(repos *repo.Repositories) *BidService {
	return &BidService{
		bidRepo:      repos.Bid,
		employeeRepo: repos.Employee,
		tenderRepo:   repos.Tender,
	}
}

func (s *BidService) CreateBid(ctx context.Context, input *entity.CreateBidInput) (*entity.BidOutputModel, error) {
	tender, err := s.tenderRepo.GetTenderById(ctx, input.TenderId)
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return nil, ErrTenderNotFound
		}

		return nil, err
	}

	if tender.Status != common.Published {
		return nil, ErrUserHasNoAccessToTender
	}

	exists, err := s.employeeRepo.DoesEmployeeExistsById(ctx, input.AuthorId)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrEmployeeNotFound
	}

	isResponsible, err := s.employeeRepo.IsEmployeeResponsible(ctx, input.AuthorId, tender.OrganizationId)
	if err != nil {
		return nil, err
	}
	if isResponsible {
		return nil, ErrBidCanNotBeProposedBySameOrganization
	}

	id, err := s.bidRepo.CreateBid(ctx, input)
	if err != nil {
		return nil, err
	}

	bid, err := s.bidRepo.GetBidById(ctx, id.String())
	if err != nil {
		return nil, err
	}

	return mapBid(bid), nil
}

// Можно ругаться, если новая версия предложения не отличается от последней
// Но в задании такого требования нет + наверно не успею это сделать, поэтому оставлю как есть
func (s *BidService) EditBidById(ctx context.Context, bidId string, username string, name string, description string) (*entity.BidOutputModel, error) {
	bid, err := s.bidRepo.GetBidById(ctx, bidId)
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return nil, ErrBidNotFound
		}

		return nil, err
	}

	employeeId, err := s.employeeRepo.GetEmployeeIdByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return nil, ErrEmployeeNotFound
		}

		return nil, err
	}

	if bid.AuthorId.String() != employeeId {
		return nil, ErrUserHasNoAccessToBid
	}

	err = s.bidRepo.EditBidById(ctx, bidId, name, description)
	if err != nil {
		return nil, err
	}

	bid, err = s.bidRepo.GetBidById(ctx, bidId)
	if err != nil {
		return nil, err
	}

	return mapBid(bid), nil
}

// Бид вне зависимости от его статуса доступен только автору и ответсвенным за организацию
func (s *BidService) GetBidStatusById(ctx context.Context, bidId string, username string) (string, error) {
	employeeId, err := s.employeeRepo.GetEmployeeIdByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return "", ErrEmployeeNotFound
		}

		return "", err
	}

	bid, err := s.bidRepo.GetBidById(ctx, bidId)
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return "", ErrBidNotFound
		}

		return "", err
	}

	if employeeId == bid.AuthorId.String() {
		return bid.Status, nil
	}

	// к нам обратился не автор бида. Значит этот пользователь должен быть из организации при тендере
	tender, err := s.tenderRepo.GetTenderById(ctx, bid.TenderId.String())
	if err != nil {
		return "", err
	}

	isEmployeeResponsible, err := s.employeeRepo.IsEmployeeResponsible(ctx, employeeId, tender.OrganizationId)
	if err != nil {
		return "", err
	}
	if !isEmployeeResponsible {
		return "", ErrUserHasNoAccessToBid
	}

	return bid.Status, nil
}

func (s *BidService) UpdateBidStatusById(ctx context.Context, bidId string, newStatus string, username string) (*entity.BidOutputModel, error) {
	bid, err := s.bidRepo.GetBidById(ctx, bidId)
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return nil, ErrBidNotFound
		}

		return nil, err
	}

	employeeId, err := s.employeeRepo.GetEmployeeIdByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return nil, ErrEmployeeNotFound
		}

		return nil, err
	}

	if bid.AuthorId.String() != employeeId {
		return nil, ErrUserHasNoAccessToBid
	}

	err = s.bidRepo.UpdateBidStatusById(ctx, bidId, newStatus)
	if err != nil {
		return nil, err
	}

	bid, err = s.bidRepo.GetBidById(ctx, bidId)
	if err != nil {
		return nil, err
	}

	return mapBid(bid), nil
}

func (s *BidService) GetBidsForTenderById(ctx context.Context, tenderId string, pg *entity.PaginationInput, username string) ([]entity.BidOutputModel, error) {
	tender, err := s.tenderRepo.GetTenderById(ctx, tenderId)
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return nil, ErrTenderNotFound
		}

		return nil, err
	}

	employeeId, err := s.employeeRepo.GetEmployeeIdByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return nil, ErrEmployeeNotFound
		}

		return nil, err
	}

	isResponsible, err := s.employeeRepo.IsEmployeeResponsible(ctx, employeeId, tender.OrganizationId)
	if err != nil {
		return nil, err
	}
	if !isResponsible {
		return nil, ErrUserHasNoAccessToTender
	}

	bids, err := s.bidRepo.GetTenderBids(ctx, tenderId, pg)
	if err != nil {
		return nil, err
	}

	return mapBids(bids), nil
}

func (s *BidService) GetUserBids(ctx context.Context, username string, pg *entity.PaginationInput) ([]entity.BidOutputModel, error) {
	employeeId, err := s.employeeRepo.GetEmployeeIdByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return nil, ErrEmployeeNotFound
		}

		return nil, err
	}

	bids, err := s.bidRepo.GetUserBids(ctx, employeeId, pg)
	if err != nil {
		return nil, err
	}

	return mapBids(bids), nil
}

func (s *BidService) SubmitBidDecision(ctx context.Context, bidId string, decision string, username string) (*entity.BidOutputModel, error) {
	employeeId, err := s.employeeRepo.GetEmployeeIdByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return nil, ErrEmployeeNotFound
		}

		return nil, err
	}

	bid, err := s.bidRepo.GetBidById(ctx, bidId)
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return nil, ErrBidNotFound
		}

		return nil, err
	}

	tender, err := s.tenderRepo.GetTenderById(ctx, bid.TenderId.String())
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return nil, ErrTenderNotFound
		}

		return nil, err
	}

	if bid.AuthorId.String() == employeeId {
		return nil, ErrBidAuthorCanNotMakeDecisionsOnIt
	}

	isResponsible, err := s.employeeRepo.IsEmployeeResponsible(ctx, employeeId, tender.OrganizationId)
	if err != nil {
		return nil, err
	}
	if !isResponsible {
		return nil, ErrUserHasNoAccessToTender
	}

	alreadySendDecision, err := s.bidRepo.AlreadySubmitApprove(ctx, bidId, employeeId)
	if err != nil {
		return nil, err
	}
	if alreadySendDecision {
		return nil, ErrAlreadyApproveBid
	}

	result := mapBid(bid)
	if bid.Decision == decision || bid.Decision == common.RejectedDecision && decision == common.ApprovedDecision {
		return result, nil
	}

	err = s.bidRepo.SubmitBidDecision(ctx, bidId, decision, employeeId, tender.OrganizationId)
	if err != nil {
		return nil, err
	}

	bid, err = s.bidRepo.GetBidById(ctx, bidId)
	if err != nil {
		return nil, err
	}

	result = mapBid(bid)

	return result, nil
}

func (s *BidService) RollbackBidVersion(ctx context.Context, bidId string, version int, username string) (*entity.BidOutputModel, error) {
	employeeId, err := s.employeeRepo.GetEmployeeIdByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return nil, ErrEmployeeNotFound
		}

		return nil, err
	}

	bid, err := s.bidRepo.GetBidById(ctx, bidId)
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return nil, ErrBidNotFound
		}

		return nil, err
	}

	if bid.AuthorId.String() != employeeId {
		return nil, ErrUserHasNoAccessToBid
	}

	err = s.bidRepo.RollbackBidVersion(ctx, bidId, version)
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return nil, ErrNoSuchVersion
		}

		return nil, err
	}

	bid, err = s.bidRepo.GetBidById(ctx, bidId)
	if err != nil {
		return nil, err
	}

	return mapBid(bid), nil
}

func (s *BidService) GetReviewsOnBidAuthorBids(ctx context.Context, tenderId string, authorUsername string, requesterUsername string, pg *entity.PaginationInput) ([]entity.ReviewOutputModel, error) {
	bidAuthorId, err := s.employeeRepo.GetEmployeeIdByUsername(ctx, authorUsername)
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return nil, ErrBidAuthorNotAnEmployee
		}

		return nil, err
	}

	requesterEmployeeId, err := s.employeeRepo.GetEmployeeIdByUsername(ctx, requesterUsername)
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return nil, ErrRequesterNotAnEmployee
		}

		return nil, err
	}

	tender, err := s.tenderRepo.GetTenderById(ctx, tenderId)
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return nil, ErrTenderNotFound
		}

		return nil, err
	}

	isResponsible, err := s.employeeRepo.IsEmployeeResponsible(ctx, requesterEmployeeId, tender.OrganizationId)
	if err != nil {
		return nil, err
	}
	if !isResponsible {
		return nil, ErrUserHasNoAccessToTender
	}

	reviews, err := s.bidRepo.GetReviewsByReceiverId(ctx, bidAuthorId, pg)
	if err != nil {
		return nil, err
	}

	return mapReviews(reviews), nil
}

func (s *BidService) SubmitBidFeedback(ctx context.Context, bidId string, username string, content string) (*entity.BidOutputModel, error) {
	employeeId, err := s.employeeRepo.GetEmployeeIdByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return nil, ErrEmployeeNotFound
		}

		return nil, err
	}

	bid, err := s.bidRepo.GetBidById(ctx, bidId)
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return nil, ErrBidNotFound
		}

		return nil, err
	}

	tender, err := s.tenderRepo.GetTenderById(ctx, bid.TenderId.String())
	if err != nil {
		return nil, err
	}

	isResponsible, err := s.employeeRepo.IsEmployeeResponsible(ctx, employeeId, tender.OrganizationId)
	if err != nil {
		return nil, err
	}
	if !isResponsible {
		return nil, ErrUserHasNoAccessToBid
	}

	receiverId := bid.AuthorId
	senderId, err := uuid.Parse(employeeId)
	if err != nil {
		return nil, err
	}

	if err = s.bidRepo.SubmitBidFeedBack(ctx, bidId, senderId, receiverId, content); err != nil {
		return nil, err
	}

	return mapBid(bid), nil
}
