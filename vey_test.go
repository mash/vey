package vey

import (
	"testing"
	"time"
)

func TestMemKeys(t *testing.T) {
	salt := []byte("salt")
	VeyTest(t, NewVey(NewDigester(salt), NewMemCache(time.Second), NewMemStore()))
}
