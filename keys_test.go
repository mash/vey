package vey

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"testing"
)

func TestMemKeys(t *testing.T) {
	salt := []byte("salt")
	k := NewVey(NewDigester(salt), NewMemCache(), NewMemStore())

	keys, err := k.GetKeys("test@example.com")
	if err != nil {
		t.Fatalf("GetKeys: %v", err)
	}
	if e, g := 0, len(keys); e != g {
		t.Errorf("len(keys) expected %v but got %v", e, g)
	}

	challenge, err := k.BeginPut("test@example.com")
	if err != nil {
		t.Fatalf("BeginPut: %v", err)
	}
	if len(challenge) == 0 {
		t.Fatalf("challenge is empty")
	}

	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	signature := ed25519.Sign(private, challenge)
	if err := k.CommitPut(challenge, signature, PublicKey{Type: SSHEd25519, PublicKey: public}); err != nil {
		t.Fatalf("CommitPut: %v", err)
	}

	invalidSignature := ed25519.Sign(private, []byte(string(challenge)+"invalid"))
	err = k.CommitPut(challenge, invalidSignature, PublicKey{Type: SSHEd25519, PublicKey: public})
	if e, g := ErrVerifyFailed, err; e != g {
		t.Fatalf("CommitPut: expected %v but got %v", e, g)
	}

	keys, err = k.GetKeys("test@example.com")
	if err != nil {
		t.Fatalf("GetKeys: %v", err)
	}
	if e, g := 1, len(keys); e != g {
		t.Fatalf("len(keys) expected %v but got %v", e, g)
	}
	if e, g := SSHEd25519, keys[0].Type; e != g {
		t.Fatalf("public key type expected %v but got %v", e, g)
	}
	if e, g := []byte(public), keys[0].PublicKey; !bytes.Equal(e, g) {
		t.Fatalf("public key expected %v but got %v", e, g)
	}
}
