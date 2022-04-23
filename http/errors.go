package http

import (
	"errors"
	"net/http"

	"github.com/mash/vey"
)

type Error struct {
	Code int    `json:"-"`
	Msg  string `json:"message"`
	Err  error  `json:"-"`
}

// Error implements error interface.
func (e Error) Error() string {
	return e.Msg
}

func (e Error) Unwrap() error {
	return e.Err
}

func NewError(err error) Error {
	var er Error
	if errors.As(err, &er) {
		return er
	}

	switch err {
	case vey.ErrInvalidEmail:
		return Error{
			Code: http.StatusBadRequest,
			Msg:  err.Error(),
			Err:  nil,
		}
	case vey.ErrNotFound:
		return Error{
			Code: http.StatusNotFound,
			Msg:  err.Error(),
			Err:  nil,
		}
	case vey.ErrVerifyFailed:
		return Error{
			Code: http.StatusBadRequest,
			Msg:  err.Error(),
			Err:  nil,
		}
	default:
		return Error{
			Code: http.StatusInternalServerError,
			Msg:  http.StatusText(http.StatusInternalServerError),
			Err:  err,
		}
	}
}

type ClientError struct {
	Msg string         // Error message
	Res *http.Response // The *http.Response returned from http.Client if it was returned from http.Client
	Err error          // underlying error if any
}

// ClientError implements error interface.
func (e ClientError) Error() string {
	return e.Msg
}

func (e ClientError) Unwrap() error {
	return e.Err
}
