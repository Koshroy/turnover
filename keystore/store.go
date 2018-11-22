package keystore

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

// Store holds public and private keys for use with the relay
type Store struct {
	privKeyPath, pubKeyPath string
	pubKey                  *rsa.PublicKey
	privKey                 *rsa.PrivateKey
}

// PubKey returns the public key of the Store
func (s *Store) PubKey() *rsa.PublicKey {
	return s.pubKey
}


// PrivKey returns the private key of the Store
func (s *Store) PrivKey() *rsa.PrivateKey {
	return s.privKey
}

// NewStore creates a new Store
func NewStore(privKeyPath, pubKeyPath string) (*Store, error) {
	if privKeyPath == "" || pubKeyPath == "" {
		return nil, fmt.Errorf("private key and public key paths are not specified properly")
	}

	b, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		return nil, fmt.Errorf("could not read public key file")
	}

	pubKeyBase, err := x509.ParsePKIXPublicKey(b)
	if err != nil {
		return nil, fmt.Errorf("could not parse public key")
	}

	pubKey, ok := pubKeyBase.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not in RSA form")
	}

	b, err = ioutil.ReadFile(privKeyPath)
	if err != nil {
		return nil, fmt.Errorf("could not read private key file")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(b)
	if err != nil {
		return nil, fmt.Errorf("could not parse private key")
	}

	return &Store{
		privKeyPath: privKeyPath,
		pubKeyPath:  pubKeyPath,
		privKey:     privKey,
		pubKey:      pubKey,
	}, nil
}

// MockStore returns a mock Store used for tests
func MockStore() *Store {
	return &Store{
		privKeyPath: "",
		pubKeyPath: "",
		privKey: &rsa.PrivateKey{},
		pubKey: &rsa.PublicKey{},
	}
}
