package http

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

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

// clientAndEmail implements ve.Vey interface.
type clientAndEmail struct {
	Client
	*email.MemSender
}

func clientVey(c Client, s *email.MemSender) vey.Vey {
	return clientAndEmail{
		Client:    c,
		MemSender: s,
	}
}

func (c clientAndEmail) BeginDelete(email string, publicKey vey.PublicKey) ([]byte, error) {
	err := c.Client.BeginDelete(email, publicKey)
	if err != nil {
		return nil, err
	}

	token := c.MemSender.Token
	if len(token) == 0 {
		return nil, errors.New("token is empty")
	}
	b, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("base64 decode error: %v", err)
	}
	return b, nil
}

func (c clientAndEmail) BeginPut(email string, publicKey vey.PublicKey) ([]byte, error) {
	err := c.Client.BeginPut(email, publicKey)
	if err != nil {
		return nil, err
	}

	challenge := c.MemSender.Challenge
	if len(challenge) == 0 {
		return nil, errors.New("challenge is empty")
	}
	b, err := base64.StdEncoding.DecodeString(challenge)
	if err != nil {
		return nil, fmt.Errorf("base64 decode error: %v", err)
	}
	return b, nil
}

// TestServer tests the server from the client's point of view,
// and from that point of view, server + email implements the vey.Vey interface.
func TestServer(t *testing.T) {
	Log = NilLogger()

	salt := []byte("salt")
	v := vey.NewVey(vey.NewDigester(salt), vey.NewMemCache(time.Second), vey.NewMemStore())
	sender_ := email.NewMemSender()

	open, _ := url.Parse("exampleapp://open")
	h := NewHandler(v, sender_, open)
	sender := sender_.(*email.MemSender)
	l := serve(t, h)

	client := NewClient("http://" + l.Addr().String())
	vey.VeyTest(t, clientVey(client, sender))

	q := url.Values{}
	q.Set("challenge", "challenge")
	next, err := client.Open(q)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if e, g := "exampleapp://open?challenge=challenge", next; e != g {
		t.Fatalf("open expected %v but got %v", e, g)
	}
}

func TestJSON(t *testing.T) {
	in := []byte(`{"challenge":"yjIW9LgQPdvZ8BGWs3HADC8zb7yk9CnwDhm4eIxxniM=","publicKey":{"key":"c3NoLWVkMjU1MTkgQUFBQUMzTnphQzFsWkRJMU5URTVBQUFBSU5scWdsbmRod0ViRXVsZU5KckU5QVVhOTdXVThXZjlTQjR2emRZdVF1U0IKCg==","type":0},"signature":"WW6Tqd1shOXvpoW5Lp/TrM5xdTBwgVfaXB6xp6nk+5YSIaQlA0oujGvj2dNnp3PGFdZoNknCm6d9Mkl4QYkADQ=="}`)
	buf := bytes.NewBuffer(in)
	dec := json.NewDecoder(buf)
	dec.DisallowUnknownFields()

	var body Body
	if err := dec.Decode(&body); err != nil {
		t.Fatal(err)
	}
}
