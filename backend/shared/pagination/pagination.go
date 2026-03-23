package pagination

import "math"

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

type Params struct {
	Limit  int
	Offset int
}

func Normalize(page, limit int) Params {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	return Params{
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}

func TotalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(limit)))
}
