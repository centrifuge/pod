#!/bin/bash
set -e
set -x

PARENT_DIR=$(pwd)
contracts=( "anchor.AnchorRegistry" "identity.Identity" "identity.IdentityRegistry" "identity.IdentityFactory" )

usage() {
  echo "Usage: $0 [--repo git-repo] [--branch branch-name] [--target output-dir]"
	exit 1
}

while test $# -gt 0
do
    case "$1" in
        --repo) echo "option repo"
            shift
            CONTRACTS_REPO=$1
            echo "Contracts Repo: ${CONTRACTS_REPO}"
            ;;
        --branch) echo "option branch"
            shift
            BRANCH=$1
            echo "Branch: ${BRANCH}"
            ;;
        --target) echo "option target"
            shift
            TARGETDIR=$1
            echo "TARGET: ${TARGETDIR}"
            ;;
        --*) echo "bad option $1"
            usage
            ;;
        *) echo "argument $1"
            usage
            ;;
    esac
    shift
done

TARGETDIR=${TARGETDIR:-${PARENT_DIR}/centrifuge}
DEFAULT_REPO="${PARENT_DIR}/vendor/github.com/centrifuge/centrifuge-ethereum-contracts"
TEMP_DIR=$DEFAULT_REPO
BRANCH=${BRANCH:-master}

if [ "X$CONTRACTS_REPO" != "X" ]
then
  TEMP_DIR=${PARENT_DIR}/temp
  mkdir -p "$TEMP_DIR"

  git clone "${CONTRACTS_REPO}" "$TEMP_DIR"
  cd "$TEMP_DIR"
  if [ "X$BRANCH" != "Xmaster" ]
  then
    git fetch
    git checkout -b "$BRANCH" origin/"${BRANCH}"
  fi
  git branch
  cd "$PARENT_DIR"
fi

CONTRACTS_REPO=${CONTRACTS_REPO:-${DEFAULT_REPO}}

echo "Building ABI Json files"
rm -Rf "$TEMP_DIR"/build/contracts
cd "$TEMP_DIR"
npm install
truffle compile
mkdir -p abi
cd "$PARENT_DIR"

echo "Extracting ABI block into its own .abi file"
for i in "${contracts[@]}"
do
  echo "Building contracts for ${i}"
  package=$(echo -n "${i}" | awk -F'.' '{print $1}')
  contract=$(echo -n "${i}" | awk -F'.' '{print $2}')
  underscore=$(echo "${contract}" | sed -e 's/\([A-Z]\)/_\1/g' | tr '[:upper:]' '[:lower:]')
  mkdir -p "${TARGETDIR}"/"${package}"
  < "$TEMP_DIR"/build/contracts/"${contract}".json jq '.abi' > "$TEMP_DIR"/abi/"${contract}".abi
  abigen --abi "$TEMP_DIR"/abi/"${contract}".abi --pkg "${package}" --type Ethereum"${contract}"Contract --out "${TARGETDIR}"/"${package}"/ethereum"${underscore}"_contract.go
done

if [ "${TEMP_DIR}" == "${PARENT_DIR}/temp" ]
then
  rm -Rf "$TEMP_DIR"
fi
