package types

import "errors"

var ErrRecordNotFound = errors.New("record not found")

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type RequestBodyError struct {
	Message string
	Errors  []FieldError
}

func (r *RequestBodyError) Error() string {
	return r.Message
}

type UserError struct {
	Message string
}

func (u *UserError) Error() string {
	return u.Message
}

type SystemError struct {
	Message string
}

func (s *SystemError) Error() string {
	return s.Message
}

func ErrIsUserError(err error) bool {
	switch err.(type) {
	case *RequestBodyError:
		return true
	case *UserError:
		return true
	default:
		return false
	}
}
