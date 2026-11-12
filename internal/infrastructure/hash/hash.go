package hash

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const defaultCost = bcrypt.DefaultCost

type Hash interface {
	Encrypt(plain string) (string, error)
	Compare(hash, plain string) error
}

type bcryptHash struct{}

func New() Hash {
	return bcryptHash{}
}

func (bcryptHash) Encrypt(plain string) (string, error) {
	out, err := bcrypt.GenerateFromPassword([]byte(plain), defaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash: %w", err)
	}
	return string(out), nil
}

func (bcryptHash) Compare(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
