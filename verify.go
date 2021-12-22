package vey

import (
	"crypto/ed25519"
)

func NewVerifier(t PublicKeyType) Verifier {
	switch t {
	case SSHEd25519:
		return SSHEd25519Verifier{}
	default:
		panic("unknown public key type")
	}
}

// SSHEd25519Verifier implements Verifier interface.
type SSHEd25519Verifier struct{}

func (v SSHEd25519Verifier) Verify(publicKey PublicKey, signature, challenge []byte) bool {
	if publicKey.Type != SSHEd25519 {
		return false
	}
	if len(publicKey.Key) != ed25519.PublicKeySize {
		return false
	}
	return ed25519.Verify(publicKey.Key, challenge, signature)
}
