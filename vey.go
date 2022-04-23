package vey

import "net/mail"

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

func validateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	return err
}

func (k vey) GetKeys(email string) ([]PublicKey, error) {
	if err := validateEmail(email); err != nil {
		return nil, ErrInvalidEmail
	}

	digest := k.digest.Of(email)
	return k.store.Get(digest)
}

func (k vey) BeginDelete(email string, publickey PublicKey) ([]byte, error) {
	if err := validateEmail(email); err != nil {
		return nil, ErrInvalidEmail
	}

	digest := k.digest.Of(email)
	token, err := NewToken()
	if err != nil {
		return nil, err
	}
	if err := k.cache.Set(token, Cached{
		EmailDigest: digest,
		PublicKey:   publickey,
	}); err != nil {
		return nil, err
	}
	return token, nil
}

func (k vey) CommitDelete(token []byte) error {
	cached, err := k.cache.Get(token)
	if err != nil {
		return err
	}
	return k.store.Delete(cached.EmailDigest, cached.PublicKey)
}

func (k vey) BeginPut(email string) ([]byte, error) {
	if err := validateEmail(email); err != nil {
		return nil, ErrInvalidEmail
	}

	digest := k.digest.Of(email)
	challenge, err := NewChallenge()
	if err != nil {
		return nil, err
	}
	if err := k.cache.Set(challenge, Cached{
		EmailDigest: digest,
	}); err != nil {
		return nil, err
	}
	return challenge, nil
}

// CommitPut verifies the signature with the public key.
// CommitPut returns ErrVerifyFailed if the signature is invalid.
func (k vey) CommitPut(challenge, signature []byte, publickey PublicKey) error {
	cached, err := k.cache.Get(challenge)
	if err != nil {
		return err
	}
	verifier := NewVerifier(publickey.Type)
	if !verifier.Verify(publickey, signature, challenge) {
		return ErrVerifyFailed
	}
	return k.store.Put(cached.EmailDigest, publickey)
}
