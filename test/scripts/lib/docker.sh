#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

function waitForCoreVMUp {
    # When dockerd handles chaincode operation, we need to wait for it to be ready
    # (it takes time to start due to chaincode compilation).
    if [[ "${CORE_VM_ENDPOINT}" =~ http://(.*):(.*) ]]; then
        echo "Waiting for VM endpoint to listen [${BASH_REMATCH[1]}:${BASH_REMATCH[2]}]..."
        declare attempt=0
        declare host=${BASH_REMATCH[1]}
        declare port=${BASH_REMATCH[2]}

        if [ "${TEST_LOCAL}" = true ]; then
            host="localhost"
        fi

        while true; do
          if [ ${attempt} -gt 240 ]; then
            echo "VM endpoint not listening"
            exit 1
          fi

          alive=$(curl -s --head --request GET ${host}:${port}/info || true)
          if [[ "${alive}" =~ ^HTTP/(.*)200 ]]; then
            break
          fi

          sleep 0.25
          attempt=$((attempt + 1))
        done
        echo "VM endpoint is listening after ${attempt} attempts"
    fi
}