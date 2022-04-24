Vey - Email Verifying Keyserver
===============================

Vey is a Email Verifying Keyserver. It is a HTTP API server that provides APIs to get, delete and put public keys for verified email addresses.

Saving a public key requires access to the email address and the private key. You prove that you have access to the email address and the private key by receiving a challenge in email and signing it with the private key.

Deleting a key requires access to the email address. You have to be able to receive an email which includes a token.

## Goals

* Do not store emails
* Realistic authentication to put and delete your keys

## Interface

The rough idea of Vey's interface.

```go
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
	Del([]byte) error
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
	PublicKey []byte
	Type      PublicKeyType
}

// EmailDigest is a hash of an email address.
type EmailDigest string

// Digester takes an email and returns a hash of it.
type Digester interface {
	Of(email string) EmailDigest
}
```

## Vey ?

From "Verify Email keYserver". Prounance like "Hey" replacing H with V.
