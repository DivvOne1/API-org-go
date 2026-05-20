package service

import "errors"

var (
	ErrNotFound          = errors.New("resource not found")
	ErrConflict          = errors.New("resource conflict")
	ErrValidation        = errors.New("validation failed")
	ErrBadRequest        = errors.New("bad request")
	ErrSelfParent        = errors.New("department cannot be parent of itself")
	ErrCycleDetected     = errors.New("department cycle detected")
	ErrMissingReassign   = errors.New("reassign_to_department_id is required for reassign mode")
	ErrInvalidDeleteMode = errors.New("invalid delete mode")
)

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

func (e ValidationError) Is(target error) bool {
	return target == ErrValidation
}
