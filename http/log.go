package http

import (
	"errors"
	"log"
)

// Log is package global variable that holds Logger.
var Log Logger = NewLogger()

// Logger logs errors.
type Logger interface {
	// Error logs an error.
	// err can be of type Error and unwrapped to return the underlying error.
	Error(err error)
}

type logger struct{}

// NewLogger returns a new default Logger that logs to stderr.
func NewLogger() Logger {
	return logger{}
}

func (l logger) Error(err error) {
	var er Error
	if errors.As(err, &er) {
		if e := er.Unwrap(); e != nil {
			log.Printf("error: %s: %s", er.Msg, e)
			return
		}
	}
	log.Printf("error: %v", err)
}

type nilLogger struct{}

func NilLogger() Logger {
	return nilLogger{}
}

func (l nilLogger) Error(err error) {}
