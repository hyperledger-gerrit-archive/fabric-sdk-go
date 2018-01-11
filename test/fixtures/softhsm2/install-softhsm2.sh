#
# Copyright IBM Corp, SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -xe

ARCH=`uname -m`

if [ $ARCH = "s390x" ]; then
  echo "deb http://ftp.us.debian.org/debian sid main" >> /etc/apt/sources.list
fi

apt-get update && \
apt-get install -y --no-install-recommends softhsm2 curl git gcc g++ libtool libltdl-dev && \
mkdir -p /var/lib/softhsm/tokens/ && \
softhsm2-util --init-token --slot 0 --label "ForFabric" --so-pin 1234 --pin 98765432 && \
mkdir -p ${GOROOT} ${GOPATH} && \
GO_URL=https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz; \
curl -o /tmp/go.tar.gz ${GO_URL} && \
tar -xvzf /tmp/go.tar.gz -C /opt/ && \
rm -rf /tmp/go.tar.gz && \
go get ${PKCS11_TOOL_URL}
