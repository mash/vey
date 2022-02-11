package vey

import (
	"log"
)

// Log is package global variable that holds Logger.
var Log Logger = NewLogger()

// Logger logs errors.
type Logger interface {
	Error(err error)
}

type logger struct{}

// NewLogger returns a new default Logger that logs to stderr.
func NewLogger() Logger {
	return logger{}
}

func (l logger) Error(err error) {
	log.Printf("error: %v", err)
}
