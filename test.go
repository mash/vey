package vey

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"golang.org/x/crypto/ssh"
)

// v's cache should be configured to expire in 2 seconds
func VeyTest(t *testing.T, v Vey) {
	testGetKeys(t, v, "test@example.com", []PublicKey{})

	challenge := testBeginPut(t, v, "test@example.com")

	edpriv, pub := testKeygen(t)

	signature := ed25519.Sign(edpriv, challenge)
	if err := v.CommitPut(challenge, signature, PublicKey{Type: SSHEd25519, Key: pub}); err != nil {
		t.Fatalf("CommitPut: %v", err)
	}

	invalidSignature := ed25519.Sign(edpriv, []byte(string(challenge)+"invalid"))
	err := v.CommitPut(challenge, invalidSignature, PublicKey{Type: SSHEd25519, Key: pub})
	if err == nil {
		t.Fatal("CommitPut: expected ErrVerifyFailed but got nil")
	}
	if e, g := ErrVerifyFailed.Error(), err.Error(); e != g {
		t.Fatalf("CommitPut: expected %#v but got %#v", e, g)
	}

	testGetKeys(t, v, "test@example.com", []PublicKey{
		{Type: SSHEd25519, Key: pub},
	})

	// try to put again and test GetKeys does not return duplicates

	challenge2 := testBeginPut(t, v, "test@example.com")
	if bytes.Equal(challenge, challenge2) {
		t.Fatalf("challenge and challenge2 should not be the same but got: %v and %v", challenge, challenge2)
	}
	signature2 := ed25519.Sign(edpriv, challenge2)
	if err := v.CommitPut(challenge2, signature2, PublicKey{Type: SSHEd25519, Key: pub}); err != nil {
		t.Fatalf("CommitPut: %v", err)
	}

	testGetKeys(t, v, "test@example.com", []PublicKey{
		{Type: SSHEd25519, Key: pub},
	})

	token, err := v.BeginDelete("test@example.com", PublicKey{Type: SSHEd25519, Key: pub})
	if err != nil {
		t.Fatalf("BeginDelete")
	}
	if len(token) == 0 {
		t.Fatalf("token is empty")
	}

	err = v.CommitDelete(token)
	if err != nil {
		t.Fatalf("CommitDelete: %v", err)
	}

	testGetKeys(t, v, "test@example.com", []PublicKey{})
}

func testKeygen(t *testing.T) (ed25519.PrivateKey, []byte) {
	edpub, edpriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	sshpub, err := ssh.NewPublicKey(edpub)
	if err != nil {
		t.Fatalf("NewPublicKey: %v", err)
	}
	pub := ssh.MarshalAuthorizedKey(sshpub)
	return edpriv, pub
}

func testGetKeys(t *testing.T, v Vey, email string, expected []PublicKey) {
	got, err := v.GetKeys(email)
	if err != nil {
		t.Fatalf("testGetKeys: %v", err)
	}
	if got == nil {
		t.Fatalf("testGetKeys: got nil keys")
	}
	if e, g := len(expected), len(got); e != g {
		t.Errorf("testGetKeys: len(got) expected %v but got %v", e, g)
	}
	for i, e := range expected {
		if e.Type != got[i].Type {
			t.Errorf("testGetKeys: expected %v but got %v", e, got[i])
		}
		if !bytes.Equal(e.Key, got[i].Key) {
			t.Errorf("testGetKeys: expected %v but got %v", e, got[i])
		}
	}
}

func testBeginPut(t *testing.T, v Vey, email string) []byte {
	challenge, err := v.BeginPut(email)
	if err != nil {
		t.Fatalf("BeginPut: %v", err)
	}
	if len(challenge) == 0 {
		t.Fatalf("challenge is empty")
	}
	return challenge
}
