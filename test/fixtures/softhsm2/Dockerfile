#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

ARG BASE_IMAGE
ARG BASE_IMAGE_TAG

FROM ${BASE_IMAGE}:${BASE_IMAGE_TAG}

ENV GOPATH=/opt/gopath \
    GOROOT=/opt/go \
    PATH=$PATH:/opt/go/bin:/opt/gopath/bin \
    GOVER="1.12.5"

COPY test/fixtures/softhsm2/install-softhsm2.sh /tmp
COPY scripts/_go/src/pkcs11helper /opt/workspace/pkcs11helper
RUN bash /tmp/install-softhsm2.sh
CMD ["/bin/bash"]

