package vault_test

import (
	"encoding/base64"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/audit"
	"github.com/hashicorp/vault/builtin/logical/database"
	"github.com/hashicorp/vault/builtin/logical/pki"
	"github.com/hashicorp/vault/builtin/logical/transit"
	"github.com/hashicorp/vault/logical"
	vaultlib "github.com/hashicorp/vault/vault"
	"github.com/unchainio/pkg/iferr"

	log "github.com/hashicorp/go-hclog"

	auditFile "github.com/hashicorp/vault/builtin/audit/file"
	credUserpass "github.com/hashicorp/vault/builtin/credential/userpass"
	vaulthttp "github.com/hashicorp/vault/http"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/vault"
)

func testVerificationFlow(tb testing.TB, csp *vault.CryptoSuite, ski []byte) {
	key, err := csp.GetKey(ski)

	iferr.Fail(err, tb)

	spew.Dump(key)

	signature, err := csp.Sign(key, []byte("blabla"), nil)

	iferr.Fail(err, tb)

	spew.Dump(signature)

	valid, err := csp.Verify(key, signature, []byte("blabla"), nil)

	iferr.Fail(err, tb)

	spew.Dump(valid)

	if !valid {
		tb.Fatalf("Signature verification failed.")
	}
}

func testVaultCryptoSuite(tb testing.TB) (*vault.CryptoSuite, func()) {
	client, closer := testVaultServer(tb)

	csp, err := vault.NewCryptoSuite(vault.WithClient(client))

	if err != nil {
		tb.Fatalf("%+v", err)
	}

	return csp, func() {
		defer closer()
	}
}

// testVaultServer creates a test vault cluster and returns a configured API
// client and closer function.
func testVaultServer(t testing.TB) (*api.Client, func()) {
	t.Helper()

	client, _, closer := testVaultServerUnseal(t)

	err := client.Sys().Mount("transit", &api.MountInput{
		Type: "transit",
	})

	if err != nil {
		t.Fatalf("%+v", err)
	}

	err = client.Sys().Mount("kv", &api.MountInput{
		Type: "kv",
	})

	if err != nil {
		t.Fatalf("%+v", err)
	}

	return client, closer
}

// testVaultServerUnseal creates a test vault cluster and returns a configured
// API client, list of unseal keys (as strings), and a closer function.
func testVaultServerUnseal(t testing.TB) (*api.Client, []string, func()) {
	t.Helper()

	return testVaultServerCoreConfig(t, &vaultlib.CoreConfig{
		DisableMlock: true,
		DisableCache: true,
		Logger:       log.NewNullLogger(),
		CredentialBackends: map[string]logical.Factory{
			"userpass": credUserpass.Factory,
		},
		AuditBackends: map[string]audit.Factory{
			"file": auditFile.Factory,
		},
		LogicalBackends: map[string]logical.Factory{
			"database":       database.Factory,
			"generic-leased": vaultlib.LeasedPassthroughBackendFactory,
			"pki":            pki.Factory,
			"transit":        transit.Factory,
		},
	})
}

// testVaultServerCoreConfig creates a new vault cluster with the given core
// configuration. This is a lower-level test helper.
func testVaultServerCoreConfig(t testing.TB, coreConfig *vaultlib.CoreConfig) (*api.Client, []string, func()) {
	t.Helper()

	cluster := vaultlib.NewTestCluster(t, coreConfig, &vaultlib.TestClusterOptions{
		HandlerFunc: vaulthttp.Handler,
	})
	cluster.Start()

	// Make it easy to get access to the active
	core := cluster.Cores[0].Core
	vaultlib.TestWaitActive(t, core)

	// Get the client already setup for us!
	client := cluster.Cores[0].Client
	client.SetToken(cluster.RootToken)

	// Convert the unseal keys to base64 encoded, since these are how the user
	// will get them.
	unsealKeys := make([]string, len(cluster.BarrierKeys))
	for i := range unsealKeys {
		unsealKeys[i] = base64.StdEncoding.EncodeToString(cluster.BarrierKeys[i])
	}

	return client, unsealKeys, func() { defer cluster.Cleanup() }
}
