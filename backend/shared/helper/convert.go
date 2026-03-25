package helper

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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

func ToString(v any) string {
	switch n := v.(type) {
	case int64:
		return strconv.FormatInt(n, 10)
	case int32:
		return strconv.FormatInt(int64(n), 10)
	case int:
		return strconv.FormatInt(int64(n), 10)
	case float64:
		return strconv.FormatFloat(n, 'f', -1, 64)
	case string:
		return n
	default:
		return ""
	}
}

func ToJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func ToJSONMarshal(v any) (json.RawMessage, error) {
	return json.Marshal(v)
}

func ToJSONUnmarshal(b []byte, v any) error {
	return json.Unmarshal(b, v)
}

func EncodeJSON(w http.ResponseWriter, encode any) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(encode)
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func ToFloat64(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int64:
		return float64(n)
	case int32:
		return float64(n)
	case int:
		return float64(n)
	default:
		return 0
	}
}
