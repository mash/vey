package email

import (
	"log"
)

type Sender interface {
	// SendToken sends a token to the email address.
	// token is the base64 encoded form of Vey's BeginDelete func return value.
	// The email recipient should call Vey server's CommitDelete API with the token.
	SendToken(email, token string) error
	// SendChallenge sends a challenge to the email address.
	// challenge is the base64 encoded form of Vey's BeginPut func return value.
	// The email recipient should sign the challenge with it's private key, and call Vey server's CommitPut API with the challenge and signature.
	SendChallenge(email, challenge string) error
}

// MemEmail implements Sender interface to be used for testing.
type MemSender struct {
	Email, Token, Challenge string
}

func NewMemSender() Sender {
	return &MemSender{}
}

func (s *MemSender) SendToken(email, token string) error {
	s.Email = email
	s.Token = token
	return nil
}

func (s *MemSender) SendChallenge(email, challenge string) error {
	s.Email = email
	s.Challenge = challenge
	return nil
}

// LogSender implements Sender interface which logs the email, token and challenge to stderr.
type LogSender struct{}

func NewLogSender() Sender {
	return LogSender{}
}

func (s LogSender) SendToken(email, token string) error {
	log.Printf("send token: %s to email: %s", token, email)
	return nil
}

func (s LogSender) SendChallenge(email, challenge string) error {
	log.Printf("send challenge: %s to email: %s", challenge, email)
	return nil
}
