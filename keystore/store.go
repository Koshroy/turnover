package keystore

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
)

// MockPrivKey is a mock private key string used for tests
// DO NOT USE in a production application
const MockPrivKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAoprnuA7StOf/puOc3Vrx/AZ3IgOh9rLBU2ftuQRrP6aElfpT
df5FsRoAeCYTPB/ISNK46Q+MDGK8WEg1lR5X3vnDbplXn982kCYVJNXxyWuk0xKJ
2v4i2Bp08Iv2RxhxMFVzyWQxrFZhU7gMaRdjTuJFe1Fd5RjKhv1hXluRoQ6cN++q
pKxVZPGCw0AVCmuS8Tk0B1/K3byQZEowscKWRuuMyBLHOP6F/bscxwJWOEEMeMXq
mqg9gCMWvUPNiy+GlyTLpQV+8CUDAKVaINZoZ+7NHpTdO3spA88WSwsxtHFY6aUd
+KMidRGOqqg3m8jy5mN6HDC+aG3KBYJkQH6CPQIDAQABAoIBAB5Xc2en9G9nXxAA
JvQzFTZm6nIBZYaIIoTyvqwog+6znsfxlwNMeCqs5GuHB03PzGqyT2jFyudAwU5j
4wO5TsI/rtUDbhNZ7m+Fe6qM9XoVSQNN0UV46H2Uqj98jm8Dw5M2Ts3EkXRMBgs+
K6qsf45nsHlrXG70ak44F6QoyAraR+oTZrev+PxzGKgesR5MNY01xOkFfuWyHOQS
3xNhtKjnqSeCZmSkNSkHJ/DyImTCNBYBh3vFHHyH23yz6GBAPATRYWVXVJQ0dWYA
iA3qZVT5qdYoJ4HPHQtJg1/jEDoo8ijzcqgbWSmUN/8lbcmRAEx63iovtXKruHar
MyYvdq0CgYEAz/0rGOw4KbkqbdJlJDwpv0vVy+nd5PYMmLgYPHmBRY46op1yfKUH
+rc+JfMEwI5vWOA2h6sUFSGhCztBaWP9RgscgAywEXJ1KFzV5tw8cmYjKDXm1601
h7QY+MDLG3MRPu5mZSMttn1WZAWfRxzO44cFwM6qAW7kXubqu4mroNMCgYEAyCPW
s8ctQOj3L4BZdbcDgoTEQ9y6xus+9Exiyh3R/P6FCT/s2TMMNvhFzdvBZIirGlBe
CP9vJCtBNmO7EWX20LehFkLt6VhnKLwC54vPFfx1G4t2KAN9EOa8x4kRoeEjjdo4
OER4g6qU+KBGJsGyiCvy9/pGaoeBsOegj0ra5q8CgYEAnqE4fYmsTCYtdhVBjqFU
NdJg/WUhF7+RW+kMkxMYxTP1BJGQ///eVhnsDIWM2k/IHMDk1hRk/LjpWueWvArG
4OUYl5EVuDjTojUr7yeJ8rZzmfeCWHyClz2EzjQ8tHLOdHDfJ8Ps2YI+oYqoMFSI
doBEowj8IJuzEa6M2PvnKoECgYEAvSPZcNbntnMzv0ltwehuQbeU/4knXnvNZ/SU
W+xomc4zDbXC8NTkU0K4PT7T+l2KTfjrlVdIwoa6P1tq25tf8InJi48+5YotG3rq
x8YBtAZ86cYXqOL7G7DjcTLhXfm1rwYuoUZcGhpoZLqa8V+WiEf4e0+jomNjNjsA
KssUKnMCgYAZTMmaUwPj/xplINHhZ+oOcG/Z7sKP2AJxYR/pzJhDBpt8A4xMY+Od
Z5dCxOjU5TRycJZhj/YmGOXtlCmtJ4DCBcF3uWEEkQREC7nvf7XYOnG5rnanYXVm
AfNmBKueIlOhLftHl8s6ziZSmjpOZpKy27QL24oyvn+Mn+FURB1fIQ==
-----END RSA PRIVATE KEY-----`

// MockPubKey is a mock public key string used for tests
// DO NOT USE in a production application
const MockPubKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAoprnuA7StOf/puOc3Vrx
/AZ3IgOh9rLBU2ftuQRrP6aElfpTdf5FsRoAeCYTPB/ISNK46Q+MDGK8WEg1lR5X
3vnDbplXn982kCYVJNXxyWuk0xKJ2v4i2Bp08Iv2RxhxMFVzyWQxrFZhU7gMaRdj
TuJFe1Fd5RjKhv1hXluRoQ6cN++qpKxVZPGCw0AVCmuS8Tk0B1/K3byQZEowscKW
RuuMyBLHOP6F/bscxwJWOEEMeMXqmqg9gCMWvUPNiy+GlyTLpQV+8CUDAKVaINZo
Z+7NHpTdO3spA88WSwsxtHFY6aUd+KMidRGOqqg3m8jy5mN6HDC+aG3KBYJkQH6C
PQIDAQAB
-----END PUBLIC KEY-----`

// Store holds public and private keys for use with the relay
type Store struct {
	privKeyPath, pubKeyPath string
	pubKeyPem               []byte
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

// PubKeyPem returns the PEM encoded public key of the Store
func (s *Store) PubKeyPem() []byte {
	return s.pubKeyPem
}

// NewStore creates a new Store
func NewStore(privKeyPath, pubKeyPath string) (*Store, error) {
	if privKeyPath == "" || pubKeyPath == "" {
		return nil, fmt.Errorf("private key and public key paths are not specified properly")
	}

	pubKeyBytes, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		return nil, fmt.Errorf("could not read public key file")
	}

	privKeyBytes, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		return nil, fmt.Errorf("could not read private key file")
	}

	store, err := makeStore(pubKeyBytes, privKeyBytes)
	if err != nil {
		return nil, err
	}

	store.privKeyPath = privKeyPath
	store.pubKeyPath = pubKeyPath
	return store, nil
}

func makeStore(pubKeyBytes, privKeyBytes []byte) (*Store, error) {
	pubKeyBytesDec, _ := pem.Decode(pubKeyBytes)
	pubKeyBase, err := x509.ParsePKIXPublicKey(pubKeyBytesDec.Bytes)
	if err != nil {
		return nil, fmt.Errorf("could not parse public key: %v", err)
	}

	pubKey, ok := pubKeyBase.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not in RSA form")
	}

	privKeyBytesDec, _ := pem.Decode(privKeyBytes)
	privKey, err := x509.ParsePKCS1PrivateKey(privKeyBytesDec.Bytes)
	if err != nil {
		return nil, fmt.Errorf("could not parse private key")
	}

	return &Store{
		privKeyPath: "",
		pubKeyPath:  "",
		privKey:     privKey,
		pubKey:      pubKey,
		pubKeyPem:   pubKeyBytes,
	}, nil
}

// MockStore returns a mock Store used for tests
func MockStore() *Store {
	store, err := makeStore([]byte(MockPubKey), []byte(MockPrivKey))
	if err != nil {
		panic(err)
	}

	store.privKeyPath = ""
	store.pubKeyPath = ""
	return store
}
