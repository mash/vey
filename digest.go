package vey

import (
	"crypto/sha256"
)

// Digest implements Digester interface.
type Digest struct {
	salt []byte
}

func NewDigester(salt []byte) Digester {
	return &Digest{salt}
}

func (d Digest) Of(email string) EmailDigest {
	h := sha256.New()
	h.Write(d.salt)
	return EmailDigest(h.Sum([]byte(email)))
}
