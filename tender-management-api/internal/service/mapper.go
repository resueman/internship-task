package service

import (
	"tender-management-api/internal/entity"
)

func mapTender(t *entity.Tender) *entity.TenderOutputModel {
	return &entity.TenderOutputModel{
		Id:             t.Id.String(),
		Name:           t.Name,
		Description:    t.Description,
		ServiceType:    t.ServiceType,
		Status:         t.Status,
		OrganizationId: t.OrganizationId.String(),
		Version:        t.Version,
		CreatedAt:      t.CreatedAt,
	}
}

func mapTenders(t []entity.Tender) []entity.TenderOutputModel {
	s := make([]entity.TenderOutputModel, 0)
	for _, tender := range t {
		s = append(s, *mapTender(&tender))
	}

	return s
}

func mapBid(t *entity.Bid) *entity.BidOutputModel {
	return &entity.BidOutputModel{
		Id:         t.Id.String(),
		Name:       t.Name,
		Status:     t.Status,
		Version:    t.Version,
		CreatedAt:  t.CreatedAt,
		AuthorType: t.AuthorType,
		AuthorId:   t.AuthorId.String(),
	}
}

func mapBids(b []entity.Bid) []entity.BidOutputModel {
	s := make([]entity.BidOutputModel, 0)
	for _, bid := range b {
		s = append(s, *mapBid(&bid))
	}

	return s
}

func mapReview(t entity.Review) *entity.ReviewOutputModel {
	return &entity.ReviewOutputModel{
		Id:          t.Id.String(),
		Description: t.Description,
		CreatedAt:   t.CreatedAt,
	}
}

func mapReviews(review []entity.Review) []entity.ReviewOutputModel {
	s := make([]entity.ReviewOutputModel, 0)
	for _, r := range review {
		s = append(s, *mapReview(r))
	}

	return s
}
