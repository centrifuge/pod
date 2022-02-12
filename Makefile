#!/bin/bash
.PHONY: help

help: ## Show this help message.
	@echo 'usage: make [target] ...'
	@echo
	@echo 'targets:'
	@egrep '^(.+)\:\ ##\ (.+)' ${MAKEFILE_LIST} | column -t -c 2 -s ':#'

clean: ##clean all dev contracts in build folder
	@rm -rf build/centrifuge-ethereum-contracts/build
	@rm -rf build/chainbridge-deploy/cb-sol-cli/chainbridge-solidity
	@rm -rf build/ethereum-bridge-contracts/out
	@rm -rf build/privacy-enabled-erc721/out
	@rm -f localAddresses
	@rm -f profile.out
	@rm -f coverage.txt

install-deps: ## Install Dependencies
	@go mod tidy
	@go install github.com/swaggo/swag/cmd/swag
	@go install github.com/ethereum/go-ethereum/cmd/abigen
	@git submodule update --init --recursive
	@command -v golangci-lint >/dev/null 2>&1 || (curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${GOPATH}/bin v1.36.0)

lint-check: ## runs linters on go code
	@golangci-lint run --skip-dirs=build/*  --disable-all --enable=golint --enable=goimports --enable=vet --enable=nakedret \
	--enable=unused --skip-dirs=resources --skip-dirs=testingutils --timeout=2m ./...;

format-go: ## formats go code
	@golangci-lint run --disable-all --enable=goimports --fix ./...

gen-swagger: ## generates the swagger documentation
	swag init --parseDependency -g ./http/router.go -o ./http
	rm -rf ./http/docs.go ./http/swagger.yaml

gen-abi-bindings: install-deps ## Generates GO ABI Bindings
	npm install --prefix build/centrifuge-ethereum-contracts
	npm run compile --prefix build/centrifuge-ethereum-contracts
	@mkdir ./tmp
	@cat build/centrifuge-ethereum-contracts/build/contracts/Identity.json | jq '.abi' > ./tmp/id.abi
	@cat build/centrifuge-ethereum-contracts/build/contracts/IdentityFactory.json | jq '.abi' > ./tmp/idf.abi
	@abigen --abi ./tmp/id.abi --pkg ideth --type IdentityContract --out ${GOPATH}/src/github.com/centrifuge/go-centrifuge/identity/ideth/identity_contract.go
	@abigen --abi ./tmp/idf.abi --pkg ideth --type FactoryContract --out ${GOPATH}/src/github.com/centrifuge/go-centrifuge/identity/ideth/factory_contract.go
	@rm -Rf ./tmp

test?="unit cmd integration testworld"
run-tests:
	@./build/scripts/test_wrapper.sh "${test}"

install: install-deps ## Builds and Install binary
	@go install -ldflags "-X github.com/centrifuge/go-centrifuge/version.gitCommit=`git rev-parse HEAD`" ./cmd/centrifuge/...

IMAGE_NAME?=centrifugeio/go-centrifuge
BRANCH=`git rev-parse --abbrev-ref HEAD`
GIT_SHORT_COMMIT=`git rev-parse --short HEAD`
TIMESTAMP=`date -u +%Y%m%d%H`
DOCKER_TAG="${BRANCH}-${TIMESTAMP}-${GIT_SHORT_COMMIT}"

build-docker: ## Build Docker Image
	@echo "Building Docker Image"
	@docker build -t ${IMAGE_NAME}:${DOCKER_TAG} .

push-to-docker: build-docker ## push docker image to registry
	@echo "Pushing Artifacts"
	@docker tag "${IMAGE_NAME}:${DOCKER_TAG}" "${IMAGE_NAME}:latest"
	@echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin
	@docker push ${IMAGE_NAME}:latest
	@docker push ${IMAGE_NAME}:${DOCKER_TAG}

push-to-swagger:
	@./build/scripts/push_to_swagger.sh

os?=`go env GOOS`
arch?=amd64
build-binary: install-deps
	@echo "Building binary for ${os}-${arch}"
	@GOOS=${os} GOARCH=${arch} go build -ldflags "-X github.com/centrifuge/go-centrifuge/version.gitCommit=`git rev-parse HEAD`" ./cmd/centrifuge/...
	@if [ -z `git tag --points-at HEAD` ]; then\
		tar -zcf centrifuge-${os}-${arch}-${GIT_SHORT_COMMIT}.tar.gz ./centrifuge;\
	else\
		tar -zcf centrifuge-${os}-${arch}-`git tag --points-at HEAD`.tar.gz ./centrifuge;\
	fi
	@echo "Built and packed into `ls *tar.gz`"

start-local-env: clean
	@FORCE_MIGRATE=true RUN_TESTS="false" ./build/scripts/test_wrapper.sh

stop-local-env:
	@CLEANUP="true" ./build/scripts/test_wrapper.sh

ethAccountKeyPath?=./build/scripts/test-dependencies/test-ethereum/migrateAccount.json
ethAccountKey?=$(shell cat ${ethAccountKeyPath})
targetDir?=${HOME}/centrifuge/testing
identityFactory:=$(shell < ./localAddresses grep "identityFactory" | awk '{print $$2}' | tr -d '\n')
recreate_config?=false
start-local-node:
	@echo "Building node..."
	@go mod vendor
	@go build -ldflags "-X github.com/centrifuge/go-centrifuge/version.gitCommit=`git rev-parse HEAD`" ./cmd/centrifuge/...
	@if [[ ! -f "${targetDir}"/config.yaml || "${recreate_config}" == "true" ]]; then \
	  rm -rf "${targetDir}"; \
      echo "Creating local test config for the Node at ${targetDir}"; \
      ./centrifuge createconfig --accountkeypath="${ethAccountKeyPath}" \
		--ethnodeurl="http://localhost:9545" --identityFactory=${identityFactory} --targetdir="${targetDir}" \
		--centchainaddr="5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY" \
		--centchainid="0xd43593c715fdd31c61141abd04a99fd6822c8558854ccde39a5684e7a56da27d" \
		--centchainsecret="//Alice" --centchainurl="ws://localhost:9944" --network=testing; \
	fi
	@echo "Starting centrifuge node..."
	@CENT_ETHEREUM_ACCOUNTS_MAIN_KEY='${ethAccountKey}' CENT_ETHEREUM_ACCOUNTS_MAIN_PASSWORD="" ./centrifuge run -c "${targetDir}"/config.yaml
