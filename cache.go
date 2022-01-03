package vey

import (
	"encoding/base64"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// MemCache implements Cache interface.
// MemCache is for testing purposes only.
// MemCache lacks expiry.
type MemCache struct {
	m      sync.Mutex
	values map[string]Cached
}

func NewMemCache() Cache {
	return &MemCache{
		values: make(map[string]Cached),
	}
}

func (c *MemCache) Set(key []byte, val Cached) error {
	c.m.Lock()
	defer c.m.Unlock()
	str := base64.StdEncoding.EncodeToString(key)
	c.values[str] = val
	return nil
}

func (c *MemCache) Get(key []byte) (Cached, error) {
	c.m.Lock()
	defer c.m.Unlock()
	str := base64.StdEncoding.EncodeToString(key)
	return c.values[str], nil
}

type DynamoDbCache struct {
	TableName string
	D         *dynamodb.DynamoDB
	expiresIn time.Duration
}

// DynamoDbCacheItem represents a single item in the DynamoDB cache table.
type DynamoDbCacheItem struct {
	ID     []byte
	Cached Cached
	// ExpiresAt is used by DynamoDB TTL to expire the item after DynamoDbCache.expiresIn duration.
	ExpiresAt time.Time `dynamodbav:",unixtime"`
}

// NewDynamoDbCache creates a new Cache implementation that is backed by DynamoDB.
// expiresIn is the duration after which the item expires, using DynamoDB TTL.
func NewDynamoDbCache(tableName string, svc *dynamodb.DynamoDB, expiresIn time.Duration) Cache {
	return &DynamoDbCache{
		TableName: tableName,
		D:         svc,
		expiresIn: expiresIn,
	}
}

func (s *DynamoDbCache) Get(b []byte) (Cached, error) {
	k, err := dynamodbattribute.MarshalMap(map[string][]byte{
		"ID": b,
	})
	if err != nil {
		return Cached{}, err
	}
	input := &dynamodb.GetItemInput{
		TableName: aws.String(s.TableName),
		Key:       k,
	}
	result, err := s.D.GetItem(input)
	if err != nil {
		log.Printf("Get: %#v, input: %#v", err, input)
		return Cached{}, err
	}
	if result.Item == nil {
		return Cached{}, nil
	}
	var item DynamoDbCacheItem
	if err := dynamodbattribute.UnmarshalMap(result.Item, &item); err != nil {
		return Cached{}, err
	}
	// https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/howitworks-ttl.html
	// "Items that have expired, but haven’t yet been deleted by TTL, still appear in reads"
	// We check if the item has expired.
	if time.Now().After(item.ExpiresAt) {
		return Cached{}, nil
	}
	return item.Cached, nil
}

// Set caches the value for the key.
func (s *DynamoDbCache) Set(b []byte, cached Cached) error {
	item := DynamoDbCacheItem{
		ID:        b,
		Cached:    cached,
		ExpiresAt: time.Now().Add(s.expiresIn),
	}
	i, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return err
	}
	input := &dynamodb.PutItemInput{
		TableName:           aws.String(s.TableName),
		Item:                i,
		ConditionExpression: aws.String("attribute_not_exists(ID)"),
	}
	_, err = s.D.PutItem(input)
	if err != nil {
		return err
	}
	return nil
}
