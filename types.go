package vey

import "bytes"

// Vey represent the public API of Email Verifying Keyserver.
// Structs that implement Vey interface may use Cache, Verifier, Store interface to implement the API.
type Vey interface {
	GetKeys(email string) ([]PublicKey, error)
	BeginDelete(email string, publicKey PublicKey) (token []byte, err error)
	CommitDelete(token []byte) error
	BeginPut(email string, publicKey PublicKey) (challenge []byte, err error)
	CommitPut(challenge, signature []byte) error
}

// Cache is a short-term key value store.
type Cache interface {
	Set([]byte, Cached) error
	Get([]byte) (Cached, error)
	Del([]byte) error
}

type Cached struct {
	EmailDigest
	PublicKey
}

// Verifier verifies the signature with the public key.
type Verifier interface {
	Verify(publicKey PublicKey, signature, challenge []byte) bool
}

// Store stores a unique set of public keys for a given email address hash.
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
	// Key is in OpenSSH authorized_keys format.
	// SSHEd25519 is only supported now, so Key should start with "ssh-ed25519 ".
	Key  []byte        `json:"key"`
	Type PublicKeyType `json:"type"`
}

// EmailDigest is a hash of an email address.
// EmailDigest is a []byte, it cannot be used as a map key.
type EmailDigest []byte

// Digester takes an email and returns a hash of it.
type Digester interface {
	Of(email string) EmailDigest
}

// Equal reports whether priv and x have the same value.
func (pub PublicKey) Equal(x PublicKey) bool {
	return pub.Type == x.Type && bytes.Equal(pub.Key, x.Key)
}
