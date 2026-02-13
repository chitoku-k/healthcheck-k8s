package service

import "errors"

type NotFoundError struct {
	Err error
}

func (e *NotFoundError) Error() string {
	return "not found: " + e.Err.Error()
}

func (e *NotFoundError) Unwrap() error {
	return e.Err
}

func NewNotFoundError(err error) error {
	return &NotFoundError{err}
}

type TimeoutError struct {
	Err error
}

func (e *TimeoutError) Error() string {
	return "timeout: " + e.Err.Error()
}

func (e *TimeoutError) Unwrap() error {
	return e.Err
}

func NewTimeoutError(err error) error {
	return &TimeoutError{err}
}

func IsNotFound(err error) bool {
	_, ok := errors.AsType[*NotFoundError](err)
	return ok
}

func IsTimeout(err error) bool {
	_, ok := errors.AsType[*TimeoutError](err)
	return ok
}
