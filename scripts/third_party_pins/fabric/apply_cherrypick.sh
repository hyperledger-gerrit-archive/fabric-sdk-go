#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

echo "Applying cherry picks (channel event client) ..."
cd $TMP_PROJECT_PATH
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/75/12375/32 && git cherry-pick FETCH_HEAD
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/77/12377/32 && git cherry-pick FETCH_HEAD
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/79/12379/32 && git cherry-pick FETCH_HEAD
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/81/12381/33 && git cherry-pick FETCH_HEAD
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/83/12483/26 && git cherry-pick FETCH_HEAD
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/63/13663/8 && git cherry-pick FETCH_HEAD
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/09/12609/26 && git cherry-pick FETCH_HEAD
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/01/13001/21 && git cherry-pick FETCH_HEAD