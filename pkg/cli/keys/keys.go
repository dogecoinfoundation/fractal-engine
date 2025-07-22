package keys

import (
	"log"

	"github.com/zalando/go-keyring"
)

type SecureStore struct {
	service string
}

func NewSecureStore() *SecureStore {
	return &SecureStore{service: "fractalengine"}
}

func (s *SecureStore) Save(key string, value string) error {
	err := keyring.Set(s.service, key, value)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func (s *SecureStore) Get(key string) (string, error) {
	secret, err := keyring.Get(s.service, key)
	if err != nil {
		log.Fatal(err)
	}

	return secret, nil
}
