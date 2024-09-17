package entity

import (
	"github.com/google/uuid"
)

type Bid struct {
	Id          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Status      string    `json:"status" db:"status"`
	TenderId    uuid.UUID `json:"tenderId" db:"tender_id"`
	AuthorType  string    `json:"authorType" db:"author_type"`
	AuthorId    uuid.UUID `json:"authorId" db:"author_id"`
	Version     int       `json:"version" db:"version"`
	CreatedAt   string    `json:"createdAt" db:"created_at"`
	Decision    string    `json:"decision" db:"decision"`
}

// service + repo input model
type CreateBidInput struct {
	Name        string // given
	Description string // given
	TenderId    string // given
	AuthorId    string // given
	AuthorType  string // given
	Status      string // should be set: "Created"
	Version     int    // should be set: 1
	// Id UUID sets automatically
	// Created_at sets automatically
}

// controller model
type BidOutputModel struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	AuthorType string `json:"authorType"`
	AuthorId   string `json:"authorId"`
	Version    int    `json:"version"`
	CreatedAt  string `json:"createdAt,"`
}
