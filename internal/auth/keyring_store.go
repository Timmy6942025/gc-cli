package auth

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
)

const defaultKeyringService = "gc-cli"

type TokenStore interface {
	SaveToken(profile string, token *oauth2.Token) error
	LoadToken(profile string) (*oauth2.Token, error)
	DeleteToken(profile string) error
}

type KeyringTokenStore struct {
	Service string
}

func NewKeyringTokenStore() *KeyringTokenStore {
	return &KeyringTokenStore{Service: defaultKeyringService}
}

func (k *KeyringTokenStore) SaveToken(profile string, token *oauth2.Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("marshal token: %w", err)
	}
	if err := keyring.Set(k.Service, profile, string(data)); err != nil {
		return fmt.Errorf("keyring set token: %w", err)
	}
	return nil
}

func (k *KeyringTokenStore) LoadToken(profile string) (*oauth2.Token, error) {
	raw, err := keyring.Get(k.Service, profile)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil, ErrNoToken
		}
		return nil, fmt.Errorf("keyring get token: %w", err)
	}
	var tok oauth2.Token
	if err := json.Unmarshal([]byte(raw), &tok); err != nil {
		return nil, fmt.Errorf("decode token: %w", err)
	}
	return &tok, nil
}

func (k *KeyringTokenStore) DeleteToken(profile string) error {
	err := keyring.Delete(k.Service, profile)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("keyring delete token: %w", err)
	}
	return nil
}
