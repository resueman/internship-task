package service

import "tender-management-api/internal/repo"

type DiagnosticsService struct {
	diagnosticsRepo repo.Diagnostics
}

func NewDiagnosticsService(repos *repo.Repositories) *DiagnosticsService {
	return &DiagnosticsService{repos.Diagnostics}
}

func (s *DiagnosticsService) Ping() error {
	if err := s.diagnosticsRepo.Ping(); err != nil {
		return err
	}

	return nil
}
