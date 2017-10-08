#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

echo "Applying cherry picks (channel event client) ..."
cd $TMP_PROJECT_PATH
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/75/12375/34 && git cherry-pick FETCH_HEAD
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/77/12377/34 && git cherry-pick FETCH_HEAD
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/79/12379/34 && git cherry-pick FETCH_HEAD
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/81/12381/35 && git cherry-pick FETCH_HEAD
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/83/12483/28 && git cherry-pick FETCH_HEAD
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/63/13663/10 && git cherry-pick FETCH_HEAD
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/09/12609/28 && git cherry-pick FETCH_HEAD
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/01/13001/23 && git cherry-pick FETCH_HEAD
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/37/14337/3 && git cherry-pick FETCH_HEAD
