package vey

import (
	"encoding/base64"
	"fmt"
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
	// might not be the correct implementation of memory cache expiry but don't want to put it in Cached
	expires   map[string]time.Time
	expiresIn time.Duration
}

func NewMemCache(expiresIn time.Duration) Cache {
	return &MemCache{
		values:    make(map[string]Cached),
		expires:   make(map[string]time.Time),
		expiresIn: expiresIn,
	}
}

func (c *MemCache) Set(key []byte, val Cached) error {
	c.m.Lock()
	defer c.m.Unlock()
	str := base64.StdEncoding.EncodeToString(key)
	c.values[str] = val
	c.expires[str] = time.Now().Add(c.expiresIn)
	return nil
}

func (c *MemCache) Get(key []byte) (Cached, error) {
	c.m.Lock()
	defer c.m.Unlock()
	str := base64.StdEncoding.EncodeToString(key)
	if val, ok := c.values[str]; !ok {
		return Cached{}, ErrNotFound
	} else if time.Now().After(c.expires[str]) {
		return Cached{}, ErrNotFound
	} else {
		return val, nil
	}
}

func (c *MemCache) Del(key []byte) error {
	c.m.Lock()
	defer c.m.Unlock()
	str := base64.StdEncoding.EncodeToString(key)
	delete(c.values, str)
	return nil
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
		return Cached{}, fmt.Errorf("MarshalMap: %w", err)
	}
	input := &dynamodb.GetItemInput{
		TableName: aws.String(s.TableName),
		Key:       k,
	}
	result, err := s.D.GetItem(input)
	if err != nil {
		Log.Error(fmt.Errorf("GetItem: input: %v, err: %w", input, err))
		return Cached{}, fmt.Errorf("GetItem: %w", err)
	}
	if result.Item == nil {
		return Cached{}, ErrNotFound
	}
	var item DynamoDbCacheItem
	if err := dynamodbattribute.UnmarshalMap(result.Item, &item); err != nil {
		return Cached{}, fmt.Errorf("UnmarshalMap: %w", err)
	}
	// https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/howitworks-ttl.html
	// "Items that have expired, but haven’t yet been deleted by TTL, still appear in reads"
	// We check if the item has expired.
	if time.Now().After(item.ExpiresAt) {
		return Cached{}, ErrNotFound
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
		return fmt.Errorf("MarshalMap: %w", err)
	}
	input := &dynamodb.PutItemInput{
		TableName:           aws.String(s.TableName),
		Item:                i,
		ConditionExpression: aws.String("attribute_not_exists(ID)"),
	}
	_, err = s.D.PutItem(input)
	if err != nil {
		Log.Error(fmt.Errorf("PutItem: input: %v, err: %w", input, err))
		return fmt.Errorf("PutItem: %w", err)
	}
	return nil
}

func (s *DynamoDbCache) Del(b []byte) error {
	k, err := dynamodbattribute.MarshalMap(map[string][]byte{
		"ID": b,
	})
	if err != nil {
		return fmt.Errorf("MarshalMap: %w", err)
	}
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(s.TableName),
		Key:       k,
	}
	_, err = s.D.DeleteItem(input)
	if err != nil {
		Log.Error(fmt.Errorf("DeleteItem: input: %v, err: %w", input, err))
		return fmt.Errorf("DeleteItem: %w", err)
	}
	return nil
}
