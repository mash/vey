//go:build aws
// +build aws

package vey

import (
	"bytes"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func TestAWSDynamoDb(t *testing.T) {
	salt := []byte("salt")

	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		t.Fatal(err)
	}
	svc := dynamodb.New(sess)

	s := NewDynamoDbStore("teststore", svc)
	c := NewDynamoDbCache("testcache", svc, time.Second*2)

	testImpl(t, NewDigester(salt), c, s)
}

func TestAWSCacheExpires(t *testing.T) {
	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		t.Fatal(err)
	}
	svc := dynamodb.New(sess)
	c := NewDynamoDbCache("testcache", svc, time.Second*2)

	token, err := NewToken()
	if err != nil {
		t.Fatal(err)
	}

	testCacheSet(t, c, token, Cached{
		EmailDigest: EmailDigest("email"),
		PublicKey: PublicKey{
			Type: SSHEd25519,
			Key:  []byte("key"),
		},
	})
	testCacheGet(t, c, token, Cached{
		EmailDigest: EmailDigest("email"),
		PublicKey: PublicKey{
			Type: SSHEd25519,
			Key:  []byte("key"),
		},
	})

	// cached value should be expired after 2 seconds
	time.Sleep(time.Second * 3)

	testCacheGet(t, c, token, Cached{})
}

func testCacheSet(t *testing.T, c Cache, key []byte, val Cached) {
	if err := c.Set(key, val); err != nil {
		t.Fatal(err)
	}
}

func testCacheGet(t *testing.T, c Cache, key []byte, expected Cached) {
	val, err := c.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal([]byte(val.EmailDigest), []byte(expected.EmailDigest)) {
		t.Fatalf("expected %v, got %v", expected, val)
	}
	if !val.PublicKey.Equal(expected.PublicKey) {
		t.Fatalf("expected %v, got %v", expected, val)
	}
}
