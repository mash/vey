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

// TestServer tests the server similarly to the way TestMemKeys in keys_test.go but over the HTTP API.
func TestServer(t *testing.T) {
	salt := []byte("salt")
	v := vey.NewVey(vey.NewDigester(salt), vey.NewMemCache(), vey.NewMemStore())
	sender_ := email.NewMemSender()
	h := NewHandler(v, sender_)
	sender := sender_.(*email.MemSender)
	l := serve(t, h)

	client := NewClient("http://" + l.Addr().String())
	keys, err := client.GetKeys("test@example.com")
	if err != nil {
		t.Fatalf("GetKeys: %v", err)
	}
	if e, g := 0, len(keys); e != g {
		t.Errorf("len(keys) expected %v but got %v", e, g)
	}

	err = client.BeginPut("test@example.com")
	if err != nil {
		t.Fatalf("BeginPut: %v", err)
	}
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

	keys, err = client.GetKeys("test@example.com")
	if err != nil {
		t.Fatalf("GetKeys: %v", err)
	}
	if e, g := 1, len(keys); e != g {
		t.Fatalf("len(keys) expected %v but got %v", e, g)
	}
	if e, g := vey.SSHEd25519, keys[0].Type; e != g {
		t.Fatalf("public key type expected %v but got %v", e, g)
	}
	if e, g := []byte(public), keys[0].Key; !bytes.Equal(e, g) {
		t.Fatalf("public key expected %v but got %v", e, g)
	}

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

	keys, err = client.GetKeys("test@example.com")
	if err != nil {
		t.Fatalf("GetKeys: %v", err)
	}
	if e, g := 0, len(keys); e != g {
		t.Errorf("len(keys) expected %v but got %v", e, g)
	}
}
