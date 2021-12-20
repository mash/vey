package vey

// vey implements Vey interface.
type vey struct {
	digest Digester
	cache  Cache
	store  Store
}

func NewVey(digest Digester, cache Cache, store Store) Vey {
	return vey{
		digest: digest,
		cache:  cache,
		store:  store,
	}
}

func (k vey) GetKeys(email string) (publickeys []PublicKey, err error) {
	digest := k.digest.Of(email)
	return k.store.Get(digest)
}

func (k vey) BeginDelete(email string, publickey PublicKey) ([]byte, error) {
	digest := k.digest.Of(email)
	token, err := NewToken()
	if err != nil {
		return nil, err
	}
	if err := k.cache.Set(string(token), Cached{
		EmailDigest: digest,
		PublicKey:   publickey,
	}); err != nil {
		return nil, err
	}
	return token, nil
}

func (k vey) CommitDelete(token []byte) error {
	cached, err := k.cache.Get(string(token))
	if err != nil {
		return err
	}
	return k.store.Delete(cached.EmailDigest, cached.PublicKey)
}

func (k vey) BeginPut(email string) ([]byte, error) {
	digest := k.digest.Of(email)
	challenge, err := NewChallenge()
	if err != nil {
		return nil, err
	}
	if err := k.cache.Set(string(challenge), Cached{
		EmailDigest: digest,
	}); err != nil {
		return nil, err
	}
	return challenge, nil
}

func (k vey) CommitPut(challenge, signature []byte, publickey PublicKey) error {
	verifier := NewVerifier(publickey.Type)
	if !verifier.Verify(publickey, signature, challenge) {
		return ErrVerifyFailed
	}
	cached, err := k.cache.Get(string(challenge))
	if err != nil {
		return err
	}
	return k.store.Put(cached.EmailDigest, publickey)
}
