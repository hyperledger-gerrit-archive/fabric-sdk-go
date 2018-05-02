package vault_test

import (
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/vault/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/vault"
	"github.com/stretchr/testify/assert"
)

func TestTransitKeyCreationWithDevServer(t *testing.T) {
	var err error

	client, err := api.NewClient(&api.Config{
		Address: "http://127.0.0.1:8200",
	})

	assert.Nil(t, err)

	secret, err := client.Logical().Write(
		"transit/keys/user5",
		map[string]interface{}{
			"type": "rsa-2048",
		},
	)

	assert.Nil(t, err)

	secret, err = client.Logical().Read("transit/keys/user5")

	assert.Nil(t, err)

	keys := secret.Data["keys"].(map[string]interface{})
	first := keys["1"].(map[string]interface{})
	publicKey := first["public_key"].(string)

	spew.Dump(publicKey)

	block, _ := pem.Decode([]byte(publicKey))

	if block == nil {
		t.Fatalf("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)

	assert.Nil(t, err)

	//_, _ = client, err
	spew.Dump(pub)
}

func TestTransitKeyCreationWithTestServer(t *testing.T) {
	var err error
	client, closer := testVaultServer(t)
	defer closer()

	secret, err := client.Logical().Write(
		"transit/keys/user5",
		map[string]interface{}{
			"type": vault.ECDSAP256,
		},
	)

	assert.Nil(t, err)

	secret, err = client.Logical().Read(
		"transit/keys/user5",
	)

	assert.Nil(t, err)

	keys := secret.Data["keys"].(map[string]interface{})
	first := keys["1"].(map[string]interface{})
	publicKey := first["public_key"].(string)

	spew.Dump(publicKey)

	block, _ := pem.Decode([]byte(publicKey))

	if block == nil {
		t.Fatalf("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)

	assert.Nil(t, err)

	spew.Dump(pub)
}

func TestKVKeyCreationWithTestServer(t *testing.T) {
	var err error
	client, closer := testVaultServer(t)
	defer closer()

	secret, err := client.Logical().Write(
		"kv/voda",
		map[string]interface{}{
			"value": "boza",
		},
	)

	assert.Nil(t, err)

	secret, err = client.Logical().Read(
		"kv/voda",
	)

	assert.Nil(t, err)
	//
	value := secret.Data["value"].(string)

	spew.Dump(value)
}
