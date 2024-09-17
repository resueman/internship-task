package service

import "errors"

var (
	ErrTenderNotFound                            = errors.New("tender not found")
	ErrBidNotFound                               = errors.New("bid not found")
	ErrEmployeeNotFound                          = errors.New("employee not found")
	ErrUserHasNoAccessToTender                   = errors.New("user doesт't have sufficient rights to access the tender")
	ErrUserHasNoAccessToBid                      = errors.New("user doesn't have sufficient rights to access the bid")
	ErrUserNotFound                              = errors.New("user with given username not found")
	ErrUnauthorizedTryToAccessWithEmployeeRights = errors.New("try to sign in as employee")

	ErrOrganizationNotFound                  = errors.New("organization not found")
	ErrUserIsNotOrganizationResponsible      = errors.New("user isn't organization responsible")
	ErrBidCanNotBeProposedBySameOrganization = errors.New("attempt to create a bid on behalf of the organization that owns the tender")

	ErrNoNewChanges                     = errors.New("no new values")
	ErrBidAuthorCanNotMakeDecisionsOnIt = errors.New("bid author can't approve or reject bid")

	ErrBidAuthorNotAnEmployee = errors.New("no bid author employee with given username")
	ErrRequesterNotAnEmployee = errors.New("no requester employee with given username")
	ErrNoSuchVersion          = errors.New("no such version")
	ErrAlreadyApproveBid      = errors.New("can't approve bid twice")
)
