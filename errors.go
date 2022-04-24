package vey

import (
	"errors"
)

var (
	// ErrNotFound indicates that the token or challenge is not found in the cache, or it has expired.
	ErrNotFound = errors.New("not found")
	// ErrVerifyFailed indicates that the signature is invalid.
	ErrVerifyFailed = errors.New("verify failed")
	ErrInvalidEmail = errors.New("invalid email")
)

func IsNotFound(err error) bool {
	if errors.Is(err, ErrNotFound) {
		return true
	}
	// We can't import github.com/mash/vey/http because of circular dependency.
	return err.Error() == "not found"
}
