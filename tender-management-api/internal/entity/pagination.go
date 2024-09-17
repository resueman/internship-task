package entity

type PaginationInput struct {
	Limit  int
	Offset int
}

func NewPaginationInput(limit int, offset int) *PaginationInput {
	return &PaginationInput{
		Limit:  limit,
		Offset: offset,
	}
}
