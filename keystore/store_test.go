package keystore

import (
	"testing"
)

func TestKeyStore(t *testing.T) {
	store := MockStore()
	pemPubKey := string(store.PubKeyPem())

	if MockPubKey != pemPubKey {
		t.Errorf(
			"mock pem public key not correct expected: %s actual: %s",
			MockPubKey,
			pemPubKey,
		)
	}
}
