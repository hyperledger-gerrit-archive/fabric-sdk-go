/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package packager

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path"
	"testing"

	pb "github.com/hyperledger/fabric/protos/peer"
)

// Test Packager wrapper ChainCode packaging
func TestPackageGolangCC(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("error from os.Getwd %v", err)
	}
	os.Setenv("GOPATH", path.Join(pwd, "../../../test/fixtures"))

	ccPackage, err := PackageCC("github.com", pb.ChaincodeSpec_GOLANG)
	if err != nil {
		t.Fatalf("error from PackageGoLangCC %v", err)
	}

	r := bytes.NewReader(ccPackage)
	gzf, err := gzip.NewReader(r)
	if err != nil {
		t.Fatalf("error from gzip.NewReader %v", err)
	}
	tarReader := tar.NewReader(gzf)
	i := 0
	exampleccExist := false
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			t.Fatalf("error from tarReader.Next() %v", err)
		}

		if header.Name == "src/github.com/example_cc/example_cc.go" {
			exampleccExist = true
		}
		i++
	}

	if !exampleccExist {
		t.Fatalf("src/github.com/example_cc/example_cc.go not exist in tar file")
	}

}

func TestPackageBinaryCC(t *testing.T) {

	ccPackage, err := PackageCC("../../../test/fixtures/src/github.com/example_cc_binary/example_cc", pb.ChaincodeSpec_BINARY)
	if err != nil {
		t.Fatalf("error from PackageGoLangCC %v", err)
	}

	r := bytes.NewReader(ccPackage)
	gzf, err := gzip.NewReader(r)
	if err != nil {
		t.Fatalf("error from gzip.NewReader %v", err)
	}
	tarReader := tar.NewReader(gzf)
	i := 0
	exampleccExist := false
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			t.Fatalf("error from tarReader.Next() %v", err)
		}

		if header.Name == "chaincode" {
			exampleccExist = true
		}
		i++
	}

	if !exampleccExist {
		t.Fatalf("chaincode does not exist in tar file")
	}

}

// TestEmptyPackageGolangCC Test Package Go ChainCode
func TestEmptyPackageGolangCC(t *testing.T) {
	os.Setenv("GOPATH", "")

	_, err := PackageCC("", pb.ChaincodeSpec_GOLANG)
	if err == nil {
		t.Fatalf("Package Empty GoLang CC must return an error.")
	}
}

// TestEmptyPackageBinaryCC Test Package Binary ChainCode
func TestEmptyPackageBinaryCC(t *testing.T) {
	_, err := PackageCC("", pb.ChaincodeSpec_BINARY)
	if err == nil {
		t.Fatalf("Package Empty GoLang CC must return an error.")
	}
	_, err = PackageCC("../../../test/fixtures/src/github.com/example_cc/", pb.ChaincodeSpec_BINARY)
	if err == nil {
		t.Fatalf("Package Empty Binary CC must not accept go source package.")
	}
}

// Test Package Go ChainCode
func TestUndefinedPackageCC(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("error from os.Getwd %v", err)
	}
	os.Setenv("GOPATH", path.Join(pwd, "../../../test/fixtures"))

	_, err = PackageCC("github.com", pb.ChaincodeSpec_UNDEFINED)
	if err == nil {
		t.Fatalf("Undefined package UndefinedCCType GoLang CC must return an error.")
	}
}
