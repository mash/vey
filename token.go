package vey

import (
	"crypto/rand"
)

func NewToken() ([]byte, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func NewChallenge() ([]byte, error) {
	return NewToken()
}
