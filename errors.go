package vey

import "errors"

var (
	// ErrNotFound indicates that the token or challenge is not found in the cache, or it has expired.
	ErrNotFound = errors.New("not found")
	// ErrVerifyFailed indicates that the signature is invalid.
	ErrVerifyFailed = errors.New("verify failed")
)
