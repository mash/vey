package vey

import "bytes"

// Vey represent the public API of Email Verifying Keyserver.
// Structs that implement Vey interface may use Cache, Verifier, Store interface to implement the API.
type Vey interface {
	GetKeys(email string) ([]PublicKey, error)
	BeginDelete(email string, publickey PublicKey) (token []byte, err error)
	CommitDelete(token []byte) error
	BeginPut(email string) (challenge []byte, err error)
	CommitPut(challenge, signature []byte, publickey PublicKey) error
}

// Cache is a short-term key value store.
type Cache interface {
	Set([]byte, Cached) error
	Get([]byte) (Cached, error)
}

type Cached struct {
	EmailDigest
	PublicKey
}

// Verifier verifies the signature with the public key.
type Verifier interface {
	Verify(publickey PublicKey, signature, challenge []byte) bool
}

// Store stores the public keys for a given email address hash.
// We do not have to store the email. The hash of it is enough.
type Store interface {
	Get(EmailDigest) ([]PublicKey, error)
	Delete(EmailDigest, PublicKey) error
	Put(EmailDigest, PublicKey) error
}

type PublicKeyType int

const (
	SSHEd25519 PublicKeyType = iota
)

type PublicKey struct {
	Key  []byte        `json:"key"`
	Type PublicKeyType `json:"type"`
}

// EmailDigest is a hash of an email address.
type EmailDigest string

// Digester takes an email and returns a hash of it.
type Digester interface {
	Of(email string) EmailDigest
}

// Equal reports whether priv and x have the same value.
func (pub PublicKey) Equal(x PublicKey) bool {
	return pub.Type == x.Type && bytes.Equal(pub.Key, x.Key)
}
