package vey

import (
	"testing"
)

func TestMemKeys(t *testing.T) {
	salt := []byte("salt")
	VeyTest(t, NewVey(NewDigester(salt), NewMemCache(), NewMemStore()))
}
