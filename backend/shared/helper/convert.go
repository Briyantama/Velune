package helper

import (
	"errors"

	"github.com/jackc/pgx/v5/pgtype"
)

func ToInt64(v any) (int64, error) {
	switch n := v.(type) {
	case int64:
		return n, nil
	case int32:
		return int64(n), nil
	case int:
		return int64(n), nil
	case float64:
		return int64(n), nil
	case pgtype.Int8:
		if n.Valid {
			return n.Int64, nil
		}
		return 0, nil
	default:
		return 0, errors.New("unexpected numeric type")
	}
}