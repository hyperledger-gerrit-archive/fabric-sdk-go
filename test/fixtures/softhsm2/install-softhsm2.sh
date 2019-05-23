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

# ----------------------------------------------------------------
# Install Golang
# ----------------------------------------------------------------
apt-get update
apt-get install -y -qq wget
mkdir -p $GOPATH
ARCH=`uname -m | sed 's|i686|386|' | sed 's|x86_64|amd64|'`

cd /tmp
wget --quiet --no-check-certificate https://storage.googleapis.com/golang/go${GOVER}.linux-${ARCH}.tar.gz
tar -xvf go${GOVER}.linux-${ARCH}.tar.gz
mv go $GOROOT
chmod 775 $GOROOT

apt-get install -y --no-install-recommends softhsm2 curl git gcc g++ libtool libltdl-dev
mkdir -p /var/lib/softhsm/tokens/
softhsm2-util --init-token --slot 0 --label "ForFabric" --so-pin 1234 --pin 98765432

cd /opt/workspace/pkcs11helper/
go install pkcs11helper/cmd/pkcs11helper
