package http

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"net"
	"net/http"
	"testing"

	"github.com/mash/vey"
	"github.com/mash/vey/email"
)

func serve(t *testing.T, h http.Handler) net.Listener {
	l, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		l.Close()
	})
	go func() {
		_ = http.Serve(l, h)
	}()

	return l
}

func testGetKeys(t *testing.T, v Client, email string, expected []vey.PublicKey) {
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

func testBeginPut(t *testing.T, v Client, email string) {
	err := v.BeginPut(email)
	if err != nil {
		t.Fatalf("BeginPut: %v", err)
	}
}

// TestServer tests the server similarly to the way TestMemKeys in keys_test.go but over the HTTP API.
func TestServer(t *testing.T) {
	salt := []byte("salt")
	v := vey.NewVey(vey.NewDigester(salt), vey.NewMemCache(), vey.NewMemStore())
	sender_ := email.NewMemSender()
	h := NewHandler(v, sender_)
	sender := sender_.(*email.MemSender)
	l := serve(t, h)

	client := NewClient("http://" + l.Addr().String())
	testGetKeys(t, client, "test@example.com", []vey.PublicKey{})

	testBeginPut(t, client, "test@example.com")

	challenge := sender.Challenge
	if len(challenge) == 0 {
		t.Fatalf("challenge is empty")
	}
	challengeb, err := base64.StdEncoding.DecodeString(challenge)
	if err != nil {
		t.Fatalf("base64 decode error: %v", err)
	}

	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	signature := ed25519.Sign(private, challengeb)
	if err := client.CommitPut(challengeb, signature, vey.PublicKey{Type: vey.SSHEd25519, Key: public}); err != nil {
		t.Fatalf("CommitPut: %v", err)
	}

	invalidSignature := ed25519.Sign(private, []byte(string(challengeb)+"invalid"))
	err = client.CommitPut(challengeb, invalidSignature, vey.PublicKey{Type: vey.SSHEd25519, Key: public})
	if e, g := vey.ErrVerifyFailed.Error(), err.Error(); e != g {
		t.Fatalf("CommitPut: expected %v but got %v", e, g)
	}

	testGetKeys(t, client, "test@example.com", []vey.PublicKey{
		{Type: vey.SSHEd25519, Key: public},
	})

	// try to put again and test GetKeys does not return duplicates

	testBeginPut(t, client, "test@example.com")

	challenge2 := sender.Challenge
	challengeb2, err := base64.StdEncoding.DecodeString(challenge2)
	if err != nil {
		t.Fatalf("base64 decode error: %v", err)
	}
	if bytes.Equal(challengeb, challengeb2) {
		t.Fatalf("challenge and challenge2 should not be the same but got: %v and %v", challenge, challenge2)
	}
	signature2 := ed25519.Sign(private, challengeb2)
	if err := client.CommitPut(challengeb2, signature2, vey.PublicKey{Type: vey.SSHEd25519, Key: public}); err != nil {
		t.Fatalf("CommitPut: %v", err)
	}

	testGetKeys(t, client, "test@example.com", []vey.PublicKey{
		{Type: vey.SSHEd25519, Key: public},
	})

	err = client.BeginDelete("test@example.com", vey.PublicKey{Type: vey.SSHEd25519, Key: public})
	if err != nil {
		t.Fatalf("BeginDelete")
	}
	token := sender.Token
	if len(token) == 0 {
		t.Fatalf("token is empty")
	}
	tokenb, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		t.Fatalf("base64 decode error: %v", err)
	}

	err = client.CommitDelete(tokenb)
	if err != nil {
		t.Fatalf("CommitDelete: %v", err)
	}

	testGetKeys(t, client, "test@example.com", []vey.PublicKey{})
}
