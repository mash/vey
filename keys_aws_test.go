//go:build aws
// +build aws

package vey

import (
	"testing"

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

	// TODO teststore
	s := NewDynamoDbStore("teststore", svc)
	testImpl(t, NewDigester(salt), NewMemCache(), s)
}
