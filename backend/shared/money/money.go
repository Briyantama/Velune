package money

import "errors"

var ErrOverflow = errors.New("amount overflow")

type Minor int64

func Add(a, b Minor) (Minor, error) {
	s := int64(a) + int64(b)
	if (int64(a) > 0 && int64(b) > 0 && s < int64(a)) || (int64(a) < 0 && int64(b) < 0 && s > int64(a)) {
		return 0, ErrOverflow
	}
	return Minor(s), nil
}

func Sub(a, b Minor) (Minor, error) {
	return Add(a, Minor(-int64(b)))
}
