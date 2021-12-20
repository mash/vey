package vey

import "sync"

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
	return s.values[d], nil
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
	s.values[d] = append(s.values[d], publickey)
	return nil
}
