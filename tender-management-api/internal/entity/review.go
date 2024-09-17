package entity

import "github.com/google/uuid"

type Review struct {
	Id          uuid.UUID
	Description string
	CreatedAt   string
	AuthorId    uuid.UUID
	ReceiverId  uuid.UUID
	BidId       string
}

type ReviewOutputModel struct {
	Id          string `json:"id"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
}
