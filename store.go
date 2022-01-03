package vey

import (
	"errors"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type MemStore struct {
	m      sync.Mutex
	values map[EmailDigest][]PublicKey
}

func NewMemStore() Store {
	return &MemStore{
		values: make(map[EmailDigest][]PublicKey),
	}
}

func (s *MemStore) Get(d EmailDigest) ([]PublicKey, error) {
	s.m.Lock()
	defer s.m.Unlock()
	ret := s.values[d]
	if ret == nil {
		return []PublicKey{}, nil
	}
	return ret, nil
}

func (s *MemStore) Delete(d EmailDigest, publickey PublicKey) error {
	s.m.Lock()
	defer s.m.Unlock()
	for i, v := range s.values[d] {
		if v.Equal(publickey) {
			s.values[d] = append(s.values[d][:i], s.values[d][i+1:]...)
			return nil
		}
	}
	return nil
}

func (s *MemStore) Put(d EmailDigest, publickey PublicKey) error {
	s.m.Lock()
	defer s.m.Unlock()
	for _, v := range s.values[d] {
		if v.Equal(publickey) {
			return nil
		}
	}
	s.values[d] = append(s.values[d], publickey)
	return nil
}

type DynamoDbStore struct {
	TableName string
	D         *dynamodb.DynamoDB
}

// DynamoDbItem represents a single item in the DynamoDB table.
type DynamoDbItem struct {
	ID string
	// PublicKeys marshals PublicKey into []byte.
	// The first byte is the PublicKey.Type and the rest is the PublicKey.Key .
	PublicKeys [][]byte `dynamodbav:"publickeys,omitempty,binaryset"`
}

func NewDynamoDbStore(tableName string, svc *dynamodb.DynamoDB) Store {
	return &DynamoDbStore{
		TableName: tableName,
		D:         svc,
	}
}

func (item DynamoDbItem) Keys() ([]PublicKey, error) {
	ret := make([]PublicKey, len(item.PublicKeys))
	for i, b := range item.PublicKeys {
		k, err := decodeDynamoDb(b)
		if err != nil {
			return nil, err
		}
		ret[i] = k
	}
	return ret, nil
}

func (s *DynamoDbStore) Get(d EmailDigest) ([]PublicKey, error) {
	key := DynamoDbItem{
		ID: string(d),
	}
	k, err := dynamodbattribute.MarshalMap(key)
	if err != nil {
		return nil, err
	}
	input := &dynamodb.GetItemInput{
		TableName: aws.String(s.TableName),
		Key:       k,
	}
	result, err := s.D.GetItem(input)
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return []PublicKey{}, nil
	}
	var item DynamoDbItem
	if err := dynamodbattribute.UnmarshalMap(result.Item, &item); err != nil {
		return nil, err
	}
	ret, err := item.Keys()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// Delete atomically deletes the public key from the set of public keys for the email digest.
func (s *DynamoDbStore) Delete(d EmailDigest, publickey PublicKey) error {
	key := DynamoDbItem{
		ID: string(d),
	}
	k, err := dynamodbattribute.MarshalMap(key)
	if err != nil {
		return err
	}
	input := &dynamodb.UpdateItemInput{
		TableName:        aws.String(s.TableName),
		Key:              k,
		UpdateExpression: aws.String("DELETE publickeys :publickey"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":publickey": {
				BS: [][]byte{encodeDynamoDb(publickey)},
			},
		},
	}
	_, err = s.D.UpdateItem(input)
	if err != nil {
		return err
	}
	return nil
}

// Put atomically adds the public key in the set of public keys for the email digest.
func (s *DynamoDbStore) Put(d EmailDigest, publickey PublicKey) error {
	key := DynamoDbItem{
		ID: string(d),
	}
	k, err := dynamodbattribute.MarshalMap(key)
	if err != nil {
		return err
	}
	input := &dynamodb.UpdateItemInput{
		TableName:        aws.String(s.TableName),
		Key:              k,
		UpdateExpression: aws.String("ADD publickeys :publickey"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":publickey": {
				BS: [][]byte{encodeDynamoDb(publickey)},
			},
		},
	}
	_, err = s.D.UpdateItem(input)
	if err != nil {
		return err
	}
	return nil
}

func encodeDynamoDb(k PublicKey) []byte {
	ret := make([]byte, 1+len(k.Key))
	ret[0] = byte(k.Type)
	copy(ret[1:], k.Key)
	return ret
}

func decodeDynamoDb(b []byte) (PublicKey, error) {
	ret := PublicKey{
		Type: PublicKeyType(b[0]),
		Key:  b[1:],
	}
	switch ret.Type {
	case SSHEd25519:
		// ok
		break
	default:
		return PublicKey{}, errors.New("unknown public key type")
	}
	return ret, nil
}
