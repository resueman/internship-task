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

type TenderService struct {
	tenderRepo   repo.Tender
	bidRepo      repo.Bid
	employeeRepo repo.Employee
}

func NewTenderService(repos *repo.Repositories) *TenderService {
	return &TenderService{
		tenderRepo:   repos.Tender,
		bidRepo:      repos.Bid,
		employeeRepo: repos.Employee,
	}
}

func (s *TenderService) CreateTender(ctx context.Context, input *entity.CreateTenderInput) (*entity.TenderOutputModel, error) {
	employeeId, err := s.employeeRepo.GetEmployeeIdByUsername(ctx, input.CreatorUsername)
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return nil, ErrEmployeeNotFound
		}

		return nil, err
	}

	organizationExists, err := s.employeeRepo.DoesOrganizationExistById(ctx, input.OrganizationId)
	if err != nil {
		return nil, err
	}
	if !organizationExists {
		return nil, ErrOrganizationNotFound
	}

	organizationId, _ := uuid.Parse(input.OrganizationId)
	isResponsible, err := s.employeeRepo.IsEmployeeResponsible(ctx, employeeId, organizationId)
	if err != nil {
		return nil, err
	}
	if !isResponsible {
		return nil, ErrUserIsNotOrganizationResponsible
	}

	id, err := s.tenderRepo.CreateTender(ctx, input)
	if err != nil {
		return nil, err
	}

	tender, err := s.tenderRepo.GetTenderById(ctx, id.String())
	if err != nil {
		return nil, err
	}

	return mapTender(tender), nil
}

// done, может редактировать любой ответственный за организацию
func (s *TenderService) EditTenderById(ctx context.Context, tenderId string, username string, name string, description string, serviceType string) (*entity.TenderOutputModel, error) {
	if name == "" && description == "" && serviceType == "" {
		return nil, ErrNoNewChanges
	}

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

	err = s.tenderRepo.EditTenderById(ctx, tenderId, name, description, serviceType)
	if err != nil {
		return nil, err
	}

	tender, err = s.tenderRepo.GetTenderById(ctx, tenderId)
	if err != nil {
		return nil, err
	}

	return mapTender(tender), nil
}

// Тендер доступен всем только если его статус Published
func (s *TenderService) GetTenderStatusById(ctx context.Context, tenderId string, username string, usernamePassed bool) (string, error) {
	var employeeId string
	var err error
	if usernamePassed {
		employeeId, err = s.employeeRepo.GetEmployeeIdByUsername(ctx, username)
		if err != nil {
			if errors.Is(err, repo_errors.ErrNotFound) {
				return "", ErrEmployeeNotFound
			}

			return "", err
		}
	}

	tender, err := s.tenderRepo.GetTenderById(ctx, tenderId)
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return "", ErrTenderNotFound
		}

		return "", err
	}

	if tender.Status == common.Published {
		return tender.Status, nil
	}

	if !usernamePassed {
		return "", ErrUnauthorizedTryToAccessWithEmployeeRights
	}

	// Тендер не публичный, значит тот, кто запрашивает должен быть из ответственных

	isEmployeeResponsible, err := s.employeeRepo.IsEmployeeResponsible(ctx, employeeId, tender.OrganizationId)
	if err != nil {
		return "", err
	}
	if !isEmployeeResponsible {
		return "", ErrUserHasNoAccessToTender
	}

	return tender.Status, nil
}

// Обновлять статус тендера может любой ответстенный за организацию, открывшую тендер
func (s *TenderService) UpdateTenderStatusById(ctx context.Context, tenderId string, newStatus string, username string) (*entity.TenderOutputModel, error) {
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

	err = s.tenderRepo.UpdateTenderStatusById(ctx, tenderId, newStatus)
	if err != nil {
		return nil, err
	}

	tender, err = s.tenderRepo.GetTenderById(ctx, tenderId)
	if err != nil {
		return nil, err
	}

	return mapTender(tender), nil
}

func (s *TenderService) GetPublishedTenders(ctx context.Context, serviceTypes []string, pg *entity.PaginationInput) ([]entity.TenderOutputModel, error) {
	tenders, err := s.tenderRepo.GetPublishedTenders(ctx, serviceTypes, pg)
	if err != nil {
		return nil, err
	}

	return mapTenders(tenders), nil
}

func (s *TenderService) GetUserTenders(ctx context.Context, username string, usernamePassed bool, pg *entity.PaginationInput) ([]entity.TenderOutputModel, error) {
	if !usernamePassed {
		return s.GetPublishedTenders(ctx, []string{}, pg)
	}
	employeeId, err := s.employeeRepo.GetEmployeeIdByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return nil, ErrEmployeeNotFound
		}

		return nil, err
	}

	organizationId, err := s.employeeRepo.GetUserOrganizationIdByEmployeeId(ctx, employeeId)
	if err != nil {
		return nil, err
	}

	tenders, err := s.tenderRepo.GetTendersByOrganizationId(ctx, organizationId, pg)
	if err != nil {
		return nil, err
	}

	return mapTenders(tenders), nil
}

func (s *TenderService) RollbackTenderVersion(ctx context.Context, tenderId string, version int, username string) (*entity.TenderOutputModel, error) {
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

	err = s.tenderRepo.RollbackTenderVersion(ctx, tenderId, version)
	if err != nil {
		if errors.Is(err, repo_errors.ErrNotFound) {
			return nil, ErrNoSuchVersion
		}

		return nil, err
	}

	tender, err = s.tenderRepo.GetTenderById(ctx, tenderId)
	if err != nil {
		return nil, err
	}

	return mapTender(tender), nil
}
