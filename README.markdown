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
// Keys represent the public API of Email Verifying Keyserver.
// Structs that implement Keys interface may use Cache, Verifier, Store interface to implement the API.
type Keys interface {
    GetKeys(email string) (publickeys []PublicKey, error)
    BeginDelete(email string, publickey PublicKey) (token string, err error)
    CommitDelete(token string) error
    BeginPut(email string) (challenge string, err error)
    CommitPut(challenge, signature string, publickey PublicKey) error
}

// Cache is a short-term key value store.
type Cache interface {
    Set(key string, val PublicKey) error
    Get(key string) (PublicKey, error)
}

// Verifier verifies the signature with the public key.
type Verifier interface {
    Verify(publickey PublicKey, signature, challenge string) error
}

// Store stores the public keys for a given email address hash.
// We do not have to store the email. The hash of it is enough.
type Store interface {
    Get(hashofemail string) ([]PublicKey, err error)
    Delete(hashofemail string, publickey PublicKey) error
    Put(hashofemail string, publickey PublicKey) error
}

enum PublicKeyType {
    SSHEd25519
}

type PublicKey struct {
    PublicKey string
    Type PublicKeyType
}
```

## Vey ?

From "Verify Email keYserver". Prounance like "Hey" replacing H with V.
