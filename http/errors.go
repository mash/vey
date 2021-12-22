package http

import "net/http"

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
