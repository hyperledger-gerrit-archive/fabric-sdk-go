/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package certutil

import "testing"

func TestTLSConfig_Bytes(t *testing.T) {
	pPem := `-----BEGIN CERTIFICATE-----
MIICSTCCAfCgAwIBAgIRAPQIzfkrCZjcpGwVhMSKd0AwCgYIKoZIzj0EAwIwdjEL
MAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBG
cmFuY2lzY28xGTAXBgNVBAoTEG9yZzEuZXhhbXBsZS5jb20xHzAdBgNVBAMTFnRs
c2NhLm9yZzEuZXhhbXBsZS5jb20wHhcNMTcwNzI4MTQyNzIwWhcNMjcwNzI2MTQy
NzIwWjB2MQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UE
BxMNU2FuIEZyYW5jaXNjbzEZMBcGA1UEChMQb3JnMS5leGFtcGxlLmNvbTEfMB0G
A1UEAxMWdGxzY2Eub3JnMS5leGFtcGxlLmNvbTBZMBMGByqGSM49AgEGCCqGSM49
AwEHA0IABMOiG8UplWTs898zZ99+PhDHPbKjZIDHVG+zQXopw8SqNdX3NAmZUKUU
sJ8JZ3M49Jq4Ms8EHSEwQf0Ifx3ICHujXzBdMA4GA1UdDwEB/wQEAwIBpjAPBgNV
HSUECDAGBgRVHSUAMA8GA1UdEwEB/wQFMAMBAf8wKQYDVR0OBCIEID9qJz7xhZko
V842OVjxCYYQwCjPIY+5e9ORR+8pxVzcMAoGCCqGSM49BAMCA0cAMEQCIGZ+KTfS
eezqv0ml1VeQEmnAEt5sJ2RJA58+LegUYMd6AiAfEe6BKqdY03qFUgEYmtKG+3Dr
O94CDp7l2k7hMQI0zQ==
-----END CERTIFICATE-----`
	tlsConfig := TLSConfig{
		Path: "",
		Pem:  pPem,
	}

	b, e := tlsConfig.Bytes()
	if e != nil {
		t.Fatalf("error loading bytes for sample cert %s", e)
	}
	if len(b) == 0 {
		t.Fatalf("cert's Bytes() call returned empty byte array")
	}
	if len(b) != len([]byte(pPem)) {
		t.Fatalf("cert's Bytes() call returned different byte array for correct pem")
	}

	// test with empty pem
	tlsConfig.Pem = ""
	b, e = tlsConfig.Bytes()
	if e != nil {
		t.Fatalf("error loading bytes for empty pem cert %s", e)
	}
	if len(b) > 0 {
		t.Fatalf("cert's Bytes() call returned non empty byte array for empty pem")
	}

	// test with wrong pem
	tlsConfig.Pem = "wrongpemvalue"
	b, e = tlsConfig.Bytes()
	if e != nil {
		t.Fatalf("error loading bytes for wrong pem cert %s", e)
	}
	if len(b) != len([]byte("wrongpemvalue")) {
		t.Fatalf("cert's Bytes() call returned different byte array for wrong pem")
	}
}

func TestTLSConfig_TLSCert(t *testing.T) {
	tlsConfig := TLSConfig{
		Path: "../../../../test/fixtures/config/mutual_tls/client_sdk_go.pem",
		Pem:  "",
	}

	c, e := tlsConfig.TLSCert()
	if e != nil {
		t.Fatalf("error loading certificate for sample cert path %s", e)
	}
	if c == nil {
		t.Fatalf("cert's TLSCert() call returned empty certificate")
	}

	// test with wrong path
	tlsConfig.Path = "dummy/path"
	c, e = tlsConfig.TLSCert()
	if e == nil {
		t.Fatal("expected error loading certificate for wrong cert path")
	}
	if c != nil {
		t.Fatalf("cert's TLSCert() call returned non empty certificate for wrong cert path")
	}

	// test with empty path and empty pem
	tlsConfig.Path = ""
	c, e = tlsConfig.TLSCert()
	if e == nil {
		t.Fatal("expected error loading certificate for empty cert path and empty pem")
	}
	if c != nil {
		t.Fatalf("cert's TLSCert() call returned non empty certificate for wrong cert path and empty pem")
	}

	// test with both correct pem and path set
	tlsConfig.Path = "../../../../test/fixtures/config/mutual_tls/client_sdk_go.pem"
	tlsConfig.Pem = `-----BEGIN CERTIFICATE-----
MIICSTCCAfCgAwIBAgIRAPQIzfkrCZjcpGwVhMSKd0AwCgYIKoZIzj0EAwIwdjEL
MAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBG
cmFuY2lzY28xGTAXBgNVBAoTEG9yZzEuZXhhbXBsZS5jb20xHzAdBgNVBAMTFnRs
c2NhLm9yZzEuZXhhbXBsZS5jb20wHhcNMTcwNzI4MTQyNzIwWhcNMjcwNzI2MTQy
NzIwWjB2MQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UE
BxMNU2FuIEZyYW5jaXNjbzEZMBcGA1UEChMQb3JnMS5leGFtcGxlLmNvbTEfMB0G
A1UEAxMWdGxzY2Eub3JnMS5leGFtcGxlLmNvbTBZMBMGByqGSM49AgEGCCqGSM49
AwEHA0IABMOiG8UplWTs898zZ99+PhDHPbKjZIDHVG+zQXopw8SqNdX3NAmZUKUU
sJ8JZ3M49Jq4Ms8EHSEwQf0Ifx3ICHujXzBdMA4GA1UdDwEB/wQEAwIBpjAPBgNV
HSUECDAGBgRVHSUAMA8GA1UdEwEB/wQFMAMBAf8wKQYDVR0OBCIEID9qJz7xhZko
V842OVjxCYYQwCjPIY+5e9ORR+8pxVzcMAoGCCqGSM49BAMCA0cAMEQCIGZ+KTfS
eezqv0ml1VeQEmnAEt5sJ2RJA58+LegUYMd6AiAfEe6BKqdY03qFUgEYmtKG+3Dr
O94CDp7l2k7hMQI0zQ==
-----END CERTIFICATE-----`
	c, e = tlsConfig.TLSCert()
	if e != nil {
		t.Fatalf("error loading certificate for sample cert path and pem %s", e)
	}
	if c == nil {
		t.Fatalf("cert's TLSCert() call returned empty certificate")
	}
}
