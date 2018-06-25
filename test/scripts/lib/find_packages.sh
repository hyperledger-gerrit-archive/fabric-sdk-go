#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

function findPackages {
    PKGS=()
    for i in "${PKG_SRC[@]}"
    do
       PKG_LIST=`${GO_CMD} list $i/... 2> /dev/null`
       while read -r line; do
          PKGS+=("$line")
       done <<< "$PKG_LIST"
    done
}

function findChangedPackages {
    # Determine which directories have changes.
    declare CHANGED=$(git diff --name-only --diff-filter=ACMRTUXB HEAD)
    declare REMOTE_REF=$(git log -1 --pretty=format:"%d" | grep '[(].*\/' | wc -l)

    # If CHANGED is empty then there is no working directory changes: fallback to last two commits.
    # Else if REFTAGS=1 then working copy commits are even with remote: only use the working copy changes.
    # Otherwise assume that the change is amending the previous commit: use both last two commit and working copy changes.
    if [[ -z "${CHANGED}" ]] || [[ "${REMOTE_REF}" -eq 0 ]]; then
        if [[ ! -z "${CHANGED}" ]]; then
            echo "Examining last commit and working directory changes"
            CHANGED+=$'\n'
        else
            echo "Examining last commit changes"
        fi

      LAST_COMMITS=($(git log -2 --pretty=format:"%h"))
      CHANGED+=$(git diff-tree --no-commit-id --name-only --diff-filter=ACMRTUXB -r ${LAST_COMMITS[1]} ${LAST_COMMITS[0]})
    else
        echo "Examining working directory changes"
    fi

    CHANGED_PKGS=()
    while read -r line; do
        if [ "$line" != "" ]; then
            DIR=`dirname $line`
            if [ "$DIR" = "." ]; then
                CHANGED_PKG+=("$REPO")
            else
                CHANGED_PKGS+=("$REPO/$DIR")
            fi
        fi
    done <<< "$CHANGED"

    # Make result unique and filter out non-Go "packages".
    CHANGED_PKGS=($(printf "%s\n" "${CHANGED_PKGS[@]}" | sort -u | xargs ${GO_CMD} list 2> /dev/null | tr '\n' ' '))
}

function filterExcludedPackages {
    FILTERED_PKGS=()
    for pkg in "${PKGS[@]}"
    do
        for i in "${CHANGED_PKGS[@]}"
        do
            if [ "$pkg" = "$i" ]; then
              FILTERED_PKGS+=("$pkg")
            fi
        done
    done

    FILTERED_PKGS=("${FILTERED_PKGS[@]}")
}

function calcDepPackages {
    echo "Calculating package dependencies ..."

    for pkg in "${PKGS[@]}"
    do
        declare testImports=$(${GO_CMD} list -f '{{.TestImports}}' ${pkg} | tr ' ' '\n' | \
            grep "^${REPO}" | \
            grep -v "^${REPO}/vendor/" | \
            grep -v "^${REPO}/internal/github.com/" | \
            grep -v "^${REPO}/third_party/github.com/" | \
            tr '\n' ' ')

        declare pkgDeps=$(${GO_CMD} list -f '{{.Deps}}' ${pkg} ${testImports} | tr ' ' '\n' | \
            grep "^${REPO}" | \
            grep -v "^${REPO}/vendor/" | \
            grep -v "^${REPO}/internal/github.com/" | \
            grep -v "^${REPO}/third_party/github.com/" | \
            tr '\n' ' ')

        declare val=$(${GO_CMD} list ${testImports} ${pkgDeps})

        export PKGDEPS__${pkg//[-\.\/]/_}="${val}"
    done
}

function appendTestImportsPackages {
    TEST_IMPORTS_PKGS=("${FILTERED_PKGS[@]}")

    for pkg in "${PKGS[@]}"
    do
        declare val=$(${GO_CMD} list -f '{{.Deps}}' ${pkg} | tr ' ' '\n' | \
            grep "^${REPO}" | \
            grep -v "^${REPO}/vendor/" | \
            grep -v "^${REPO}/internal/github.com/" | \
            grep -v "^${REPO}/third_party/github.com/" | \
            tr '\n' ' ')

        export PKGDEPS__${pkg//[-\.\/]/_}="${val}"
    done
}

function appendDepPackages {
    calcDepPackages

    DEP_PKGS=("${FILTERED_PKGS[@]}")

    # For each changed package, see if a candidate package uses that changed package as a dependency.
    # If so, include that candidate package.
    for cpkg in "${CHANGED_PKGS[@]}"
    do
        for pkg in "${PKGS[@]}"
        do
            declare key="PKGDEPS__${pkg//[-\.\/]/_}"
            pkgDeps=(${!key})

            for i in "${pkgDeps[@]}"
            do
                if [ "$cpkg" = "$i" ]; then
                  DEP_PKGS+=("$pkg")
                fi
            done
        done
    done

    DEP_PKGS=($(printf "%s\n" "${DEP_PKGS[@]}" | sort -u | tr '\n' ' '))
}

# packagesToDirs convert packages to directories
function packagesToDirs {
    DIRS=()
    for i in "${PKGS[@]}"
    do
        pkgdir=${i#$REPO/}
        DIRS+=(${pkgdir})
    done
}