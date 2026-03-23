package repository

import "errors"

var ErrNotFound = errors.New("not found")
var ErrOptimisticLock = errors.New("optimistic lock conflict")
var ErrInsufficientBalance = errors.New("insufficient balance")
