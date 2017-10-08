#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

echo "Applying cherry picks (channel event client) ..."
cd $TMP_PROJECT_PATH


#git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/17/14617/5 && git cherry-pick FETCH_HEAD
#git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/63/14663/4 && git cherry-pick FETCH_HEAD
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/75/12375/39 && git cherry-pick FETCH_HEAD
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/77/12377/39 && git cherry-pick FETCH_HEAD
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/79/12379/39 && git cherry-pick FETCH_HEAD
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/81/12381/40 && git cherry-pick FETCH_HEAD
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/83/12483/33 && git cherry-pick FETCH_HEAD
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/63/13663/15 && git cherry-pick FETCH_HEAD
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/09/12609/33 && git cherry-pick FETCH_HEAD
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/01/13001/28 && git cherry-pick FETCH_HEAD
git fetch https://gerrit.hyperledger.org/r/fabric refs/changes/37/14337/8 && git cherry-pick FETCH_HEAD
