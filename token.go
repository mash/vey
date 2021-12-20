package vey

import (
	"crypto/rand"
	"encoding/base64"
)

func NewToken() ([]byte, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	out := make([]byte, base64.RawURLEncoding.EncodedLen(len(b)))
	base64.RawURLEncoding.Encode(out, b)
	return out, nil
}

func NewChallenge() ([]byte, error) {
	return NewToken()
}
