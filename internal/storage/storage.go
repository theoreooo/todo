package storage

import "errors"

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrURLExists    = errors.New("url exists")
)
