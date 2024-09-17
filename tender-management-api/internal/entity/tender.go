package entity

import (
	"github.com/google/uuid"
)

// db model
type Tender struct {
	Id             uuid.UUID `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	Description    string    `json:"description" db:"description"`
	ServiceType    string    `json:"serviceType" db:"service_type"`
	Status         string    `json:"status" db:"status"`
	OrganizationId uuid.UUID `json:"organizationId" db:"organization_id"`
	Version        int       `json:"version" db:"version"`
	CreatedAt      string    `json:"createdAt" db:"created_at"`
}

// service + repo input model
type CreateTenderInput struct {
	Name            string // given
	Description     string // given
	ServiceType     string // given
	OrganizationId  string // given
	CreatorUsername string // given
	Status          string // should be set: "Created"
	Version         int    // should be set: 1
	// Id UUID sets automatically
	// CreatedAt sets automatically
}

// controller model
type TenderOutputModel struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	ServiceType    string `json:"serviceType"`
	Status         string `json:"status"`
	OrganizationId string `json:"organizationId"`
	Version        int    `json:"version"`
	CreatedAt      string `json:"createdAt"`
}
