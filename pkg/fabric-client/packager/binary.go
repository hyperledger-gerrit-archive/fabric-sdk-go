/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package packager

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
)

// PackageBinaryCC ...packages binary provided as a absolute path to tar bytes
func PackageBinaryCC(chaincodePath string) ([]byte, error) {
	// set up the gzip writer
	var codePackage bytes.Buffer
	gw := gzip.NewWriter(&codePackage)
	tw := tar.NewWriter(gw)

	logger.Debugf("generateTarGz for %s", chaincodePath)
	err := packEntry(tw, gw, &Descriptor{name: "chaincode", fqp: chaincodePath}, 0100555)
	if err != nil {
		closeStream(tw, gw)
		return nil, fmt.Errorf("error from packEntry for %s error %s", chaincodePath, err.Error())
	}

	closeStream(tw, gw)

	return codePackage.Bytes(), nil
}
