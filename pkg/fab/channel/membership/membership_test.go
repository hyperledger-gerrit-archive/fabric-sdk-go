/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package membership

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
	"github.com/stretchr/testify/assert"
)

var crl []byte

func TestCertSignedWithUnknownAuthority(t *testing.T) {
	var err error
	goodMSPID := "GoodMSP"
	ctx := mocks.NewMockProviderContext()
	cfg := mocks.NewMockChannelCfg("")
	// Test good config input
	cfg.MockMsps = []*mb.MSPConfig{buildMSPConfig(goodMSPID, []byte(validRootCA))}
	m, err := New(Context{Providers: ctx}, cfg)
	assert.Nil(t, err)
	assert.NotNil(t, m)

	invalidSignatureCrt := []byte(invalidSignaturePem)

	// We serialize identities by prepending the MSPID and appending the ASN.1 DER content of the cert
	sID := &mb.SerializedIdentity{Mspid: goodMSPID, IdBytes: invalidSignatureCrt}
	goodEndorser, err := proto.Marshal(sID)
	assert.Nil(t, err)
	err = m.Validate(goodEndorser)
	if !strings.Contains(err.Error(), "certificate signed by unknown authority") {
		t.Fatalf("Expected error:'supplied identity is not valid: x509: certificate signed by unknown authority'")
	}
}

func TestRevokedCertCRLSIgnatureMismatch(t *testing.T) {
	var err error
	goodMSPID := "GoodMSP"
	ctx := mocks.NewMockProviderContext()
	cfg := mocks.NewMockChannelCfg("")
	crl, err = generateCRL(validRootCA, certPem)
	if err != nil {
		t.Fatalf("Error %v", err)
	}
	// Test good config input
	cfg.MockMsps = []*mb.MSPConfig{buildMSPConfig(goodMSPID, []byte(validRootCA))}
	m, err := New(Context{Providers: ctx}, cfg)
	assert.Nil(t, err)
	assert.NotNil(t, m)

	// We serialize identities by prepending the MSPID and appending the ASN.1 DER content of the cert
	sID := &mb.SerializedIdentity{Mspid: goodMSPID, IdBytes: []byte(certPem)}
	goodEndorser, err := proto.Marshal(sID)
	assert.Nil(t, err)
	//If CRL list is not sign proprly MSP just returns warning about signature
	err = m.Validate(goodEndorser)
	assert.Nil(t, err)
}

func TestNewMembership(t *testing.T) {
	goodMSPID := "GoodMSP"
	badMSPID := "BadMSP"

	ctx := mocks.NewMockProviderContext()
	cfg := mocks.NewMockChannelCfg("")

	// Test bad config input
	cfg.MockMsps = []*mb.MSPConfig{buildMSPConfig(goodMSPID, []byte("invalid"))}
	m, err := New(Context{Providers: ctx}, cfg)
	assert.NotNil(t, err)
	assert.Nil(t, m)

	// Test good config input
	cfg.MockMsps = []*mb.MSPConfig{buildMSPConfig(goodMSPID, []byte(validRootCA))}
	m, err = New(Context{Providers: ctx}, cfg)
	assert.Nil(t, err)
	assert.NotNil(t, m)

	// We serialize identities by prepending the MSPID and appending the ASN.1 DER content of the cert
	sID := &mb.SerializedIdentity{Mspid: goodMSPID, IdBytes: []byte(certPem)}
	goodEndorser, err := proto.Marshal(sID)
	assert.Nil(t, err)

	sID = &mb.SerializedIdentity{Mspid: badMSPID, IdBytes: []byte(certPem)}
	badEndorser, err := proto.Marshal(sID)
	assert.Nil(t, err)

	assert.Nil(t, m.Validate(goodEndorser))
	assert.NotNil(t, m.Validate(badEndorser))

	assert.Nil(t, m.Verify(goodEndorser, []byte("test"), []byte("test1")))
	assert.NotNil(t, m.Verify(badEndorser, []byte("test"), []byte("test1")))
}

func buildMSPConfig(name string, root []byte) *mb.MSPConfig {
	return &mb.MSPConfig{
		Type:   0,
		Config: marshalOrPanic(buildfabricMSPConfig(name, root)),
	}
}

func buildfabricMSPConfig(name string, root []byte) *mb.FabricMSPConfig {
	config := &mb.FabricMSPConfig{
		Name:                          name,
		Admins:                        [][]byte{},
		IntermediateCerts:             [][]byte{},
		OrganizationalUnitIdentifiers: []*mb.FabricOUIdentifier{},
		RootCerts:                     [][]byte{root},
		SigningIdentity:               nil,
	}
	if len(crl) > 0 {
		config.RevocationList = [][]byte{crl}
	}
	return config

}

func marshalOrPanic(pb proto.Message) []byte {
	data, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}
	return data
}

func generateCRL(rootCA string, certPem string) ([]byte, error) {

	block, _ := pem.Decode([]byte(rootCA))
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err

	}

	block, _ = pem.Decode([]byte(certPem))
	certR, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	ecdsaPriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	loc := time.FixedZone("Oz/Atlantis", int((2 * time.Hour).Seconds()))
	now := time.Unix(1000, 0).In(loc)
	expiry := time.Unix(10000, 0)
	revokedCerts := []pkix.RevokedCertificate{
		{
			SerialNumber:   certR.SerialNumber,
			RevocationTime: now.UTC(),
		},
	}
	crlBytes, err := cert.CreateCRL(rand.Reader, ecdsaPriv, revokedCerts, now, expiry)
	if err != nil {
		return nil, err
	}
	return crlBytes, nil
}

func TestGenerateCRL(t *testing.T) {
	block, _ := pem.Decode([]byte(orgTwoCA))
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("Error %v", err)

	}

	block, _ = pem.Decode([]byte(org2Peer1CertPEM))
	certR, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("Error %v", err)
	}
	ecdsaPriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Error %v", err)
	}

	loc := time.FixedZone("Oz/Atlantis", int((2 * time.Hour).Seconds()))
	now := time.Unix(1000, 0).In(loc)
	expiry := time.Unix(10000, 0)
	revokedCerts := []pkix.RevokedCertificate{
		{
			SerialNumber:   certR.SerialNumber,
			RevocationTime: now.UTC(),
		},
	}
	crlBytes, err := cert.CreateCRL(rand.Reader, ecdsaPriv, revokedCerts, now, expiry)
	if err != nil {
		t.Fatalf("Error %v", err)
	}
	parsedCRL, err := x509.ParseDERCRL(crlBytes)
	if err != nil {
		t.Fatalf("Error %v", err)
	}
	pemdata := pem.EncodeToMemory(
		&pem.Block{
			Type:  "X509 CRL",
			Bytes: crlBytes,
		},
	)
	if parsedCRL == nil {
		t.Fatalf("Error %v", "Expected CRL to be created")
	}

	if pemdata == nil {
		t.Fatalf("Error %v", "Expected valid PEM")
	}
	//fmt.Printf("parsedCRL %v\n\n %s\n", parsedCRL, pemdata)
}

var validRootCA = `-----BEGIN CERTIFICATE-----
MIICQzCCAemgAwIBAgIQYZpqGmcswky9Iy1SHBIm8zAKBggqhkjOPQQDAjBzMQsw
CQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZy
YW5jaXNjbzEZMBcGA1UEChMQb3JnMS5leGFtcGxlLmNvbTEcMBoGA1UEAxMTY2Eu
b3JnMS5leGFtcGxlLmNvbTAeFw0xNzA3MjgxNDI3MjBaFw0yNzA3MjYxNDI3MjBa
MHMxCzAJBgNVBAYTAlVTMRMwEQYDVQQIEwpDYWxpZm9ybmlhMRYwFAYDVQQHEw1T
YW4gRnJhbmNpc2NvMRkwFwYDVQQKExBvcmcxLmV4YW1wbGUuY29tMRwwGgYDVQQD
ExNjYS5vcmcxLmV4YW1wbGUuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE
3WtPeUzseT9Wp9VUtkx6mF84plyhgTlI2pbrHa4wYKFSoQGmrt83px6Q5Qu9EmhW
1y6Fr8DxkHvvg1NX0bCGyaNfMF0wDgYDVR0PAQH/BAQDAgGmMA8GA1UdJQQIMAYG
BFUdJQAwDwYDVR0TAQH/BAUwAwEB/zApBgNVHQ4EIgQgh5HRNj6JUV+a+gQrBpOi
xwS7jdldKPl9NUmiuePENS0wCgYIKoZIzj0EAwIDSAAwRQIhALUmxdk1FP8uL1so
nLdU8D8CS2PW5DLbaMjhR1KVK3b7AiAD5vkgX1PXPRsFFYlbkp/Y+nDdDy+mk3N7
K7xCT/QO7Q==
-----END CERTIFICATE-----
`

var certPem = `-----BEGIN CERTIFICATE-----
MIICGDCCAb+gAwIBAgIQXOaCoTss6vG3zb/vRGWXuDAKBggqhkjOPQQDAjBzMQsw
CQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZy
YW5jaXNjbzEZMBcGA1UEChMQb3JnMS5leGFtcGxlLmNvbTEcMBoGA1UEAxMTY2Eu
b3JnMS5leGFtcGxlLmNvbTAeFw0xNzA3MjgxNDI3MjBaFw0yNzA3MjYxNDI3MjBa
MFsxCzAJBgNVBAYTAlVTMRMwEQYDVQQIEwpDYWxpZm9ybmlhMRYwFAYDVQQHEw1T
YW4gRnJhbmNpc2NvMR8wHQYDVQQDExZwZWVyMC5vcmcxLmV4YW1wbGUuY29tMFkw
EwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEWXupBEBzx/Mnjz1hzIUeOGiVR4CV/7aS
Qv0aokqJanTD+x8MaavBNYbPUwwzUNc7c1Ydd12gUNHPnyj/r1YyuaNNMEswDgYD
VR0PAQH/BAQDAgeAMAwGA1UdEwEB/wQCMAAwKwYDVR0jBCQwIoAgh5HRNj6JUV+a
+gQrBpOixwS7jdldKPl9NUmiuePENS0wCgYIKoZIzj0EAwIDRwAwRAIgT2CAHCtr
Ro1YX8QuD6dSZUAOmptC+xU5xhp+2MeY2BkCIHmLOMBU5KIyJ5Rah4QeiswJ/pge
0eiDDUjXWGduFy4x
-----END CERTIFICATE-----`

var keyPem = `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQg9e5CQn0/jnFMQj9o
xs12HqzJpUa4j7Sj3spkbL+3dFGhRANCAARZe6kEQHPH8yePPWHMhR44aJVHgJX/
tpJC/RqiSolqdMP7Hwxpq8E1hs9TDDNQ1ztzVh13XaBQ0c+fKP+vVjK5
-----END PRIVATE KEY-----
`

var invalidSignaturePem = `-----BEGIN CERTIFICATE-----
MIICCzCCAbKgAwIBAgIQaiOerd7fYdLv3WOe3G7maTAKBggqhkjOPQQDAjBXMQsw
CQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZy
YW5jaXNjbzEOMAwGA1UEChMFcm9vdDIxCzAJBgNVBAMTAmNhMB4XDTE3MTIyMTE3
MTE1NFoXDTI3MTIxOTE3MTE1NFowVTELMAkGA1UEBhMCVVMxEzARBgNVBAgTCkNh
bGlmb3JuaWExFjAUBgNVBAcTDVNhbiBGcmFuY2lzY28xGTAXBgNVBAMTEHNpZ25j
ZXJ0LXJldm9rZWQwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAATkCUK/7PBlDVY6
IyYVdLIJaHjz5Bx3mTMwySYwUsDYU0zD0btx0EBAKjTMDiLqkC5dllaxrU4gzHxr
5hy99+zjo2IwYDAOBgNVHQ8BAf8EBAMCBaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEw
DAYDVR0TAQH/BAIwADArBgNVHSMEJDAigCBdC72qnK+2ajHaE61O7EwMxTJgqgm7
evx+2WCfZMfxOjAKBggqhkjOPQQDAgNHADBEAiAnGpZxlGGG4GIRc3bmrIqtG7sz
O/7VzRFysxkwySQCNwIgedom1wB4w/W/p05tdh6YXo8kLrEOWUb9KMchm3iaKT8=
-----END CERTIFICATE-----`

//use this one to sign CRL
var orgTwoCA = `-----BEGIN CERTIFICATE-----
MIICRDCCAeqgAwIBAgIRANqpQ8r//fDaj4j6kuGJv8gwCgYIKoZIzj0EAwIwczEL
MAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBG
cmFuY2lzY28xGTAXBgNVBAoTEG9yZzIuZXhhbXBsZS5jb20xHDAaBgNVBAMTE2Nh
Lm9yZzIuZXhhbXBsZS5jb20wHhcNMTcwNzI4MTQyNzIwWhcNMjcwNzI2MTQyNzIw
WjBzMQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMN
U2FuIEZyYW5jaXNjbzEZMBcGA1UEChMQb3JnMi5leGFtcGxlLmNvbTEcMBoGA1UE
AxMTY2Eub3JnMi5leGFtcGxlLmNvbTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IA
BLdwS4lO/WKTrsQt+Q2bLMIbntuM7Teg6fEXvKrpIHFNzaCsTlemFUVxQUugQfUA
/GGIaomaE1STfvbCtElCsSOjXzBdMA4GA1UdDwEB/wQEAwIBpjAPBgNVHSUECDAG
BgRVHSUAMA8GA1UdEwEB/wQFMAMBAf8wKQYDVR0OBCIEIKJZIE29atsUwFpuAt6U
Vnsqn32+nmoGO6dn1CvwtUTBMAoGCCqGSM49BAMCA0gAMEUCIQCH8+Vw0L38dv/v
9gWvLhQv69q2bS0FBiAFwR4M17Z/2QIgH5W6rmsItiwa7nD0eZyiGmCzzQXW01b4
5fDo4hNhETQ=
-----END CERTIFICATE-----`

//this one will be revoked
var org2Peer1CertPEM = `-----BEGIN CERTIFICATE-----
MIICGDCCAb+gAwIBAgIQMM7J6tIoILDsSk24evyETjAKBggqhkjOPQQDAjBzMQsw
CQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZy
YW5jaXNjbzEZMBcGA1UEChMQb3JnMi5leGFtcGxlLmNvbTEcMBoGA1UEAxMTY2Eu
b3JnMi5leGFtcGxlLmNvbTAeFw0xNzA3MjgxNDI3MjBaFw0yNzA3MjYxNDI3MjBa
MFsxCzAJBgNVBAYTAlVTMRMwEQYDVQQIEwpDYWxpZm9ybmlhMRYwFAYDVQQHEw1T
YW4gRnJhbmNpc2NvMR8wHQYDVQQDExZwZWVyMS5vcmcyLmV4YW1wbGUuY29tMFkw
EwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE2oM2fKlTXIXBLdnMXWBvnYfUm10C4mGO
Ap0OqF59S1ZOtSzu7kWpj7PD6Vhkrg3d1rdS4LlxpNdnrcRCvTxgVqNNMEswDgYD
VR0PAQH/BAQDAgeAMAwGA1UdEwEB/wQCMAAwKwYDVR0jBCQwIoAgolkgTb1q2xTA
Wm4C3pRWeyqffb6eagY7p2fUK/C1RMEwCgYIKoZIzj0EAwIDRwAwRAIgaMWMM44S
U2fCmYmPt+KJJSLV/erlEquzk4AycbmQkQwCIALAfpWBvygGuCKbJ1X8yNYtAr8c
zJYESEkHHFx/MMLl
-----END CERTIFICATE-----`
