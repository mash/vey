package vey

import (
	"crypto/ed25519"
	"errors"
	"fmt"

	"golang.org/x/crypto/ssh"
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

func (v SSHEd25519Verifier) Verify(pub PublicKey, signature, challenge []byte) bool {
	if pub.Type != SSHEd25519 {
		return false
	}
	out, _, _, _, err := ssh.ParseAuthorizedKey(pub.Key)
	if err != nil {
		Log.Error(fmt.Errorf("ParseAuthorizedKey: %w", err))
		return false
	}
	c, ok := out.(ssh.CryptoPublicKey)
	if !ok {
		Log.Error(errors.New("not a CryptoPublicKey"))
		return false
	}
	p, ok := c.CryptoPublicKey().(ed25519.PublicKey)
	if !ok {
		Log.Error(fmt.Errorf("parsed public key was not ed25519.PublicKey: %v", pub))
		return false
	}
	return ed25519.Verify(p, challenge, signature)
}
