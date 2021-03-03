#!/bin/bash
.PHONY: help

help: ## Show this help message.
	@echo 'usage: make [target] ...'
	@echo
	@echo 'targets:'
	@egrep '^(.+)\:\ ##\ (.+)' ${MAKEFILE_LIST} | column -t -c 2 -s ':#'

clean-contracts: ##clean all dev contracts in build folder
	@rm -rf build/centrifuge-ethereum-contracts/build
	@rm -rf build/chainbridge-deploy/cb-sol-cli/chainbridge-solidity
	@rm -rf build/ethereum-bridge-contracts/out
	@rm -rf build/privacy-enabled-erc721/out

clean: ##clean vendor's folder. Should be run before a make install
	@echo 'cleaning previous /vendor folder'
	@rm -rf vendor/
	@echo 'done cleaning'

install-deps: ## Install Dependencies
	@go mod tidy
	@go install github.com/jteeuwen/go-bindata/go-bindata
	@go install github.com/swaggo/swag/cmd/swag
	@go install github.com/ethereum/go-ethereum/cmd/abigen
	@go install github.com/karalabe/xgo
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

generate: ## autogenerate go files for config
	go generate ./config/configuration.go

install-subkey: ## installs subkey
	wget -P ${HOME}/.cargo/bin/ https://storage.googleapis.com/centrifuge-dev-public/subkey-rc6
	mv ${HOME}/.cargo/bin/subkey-rc6 ${HOME}/.cargo/bin/subkey
	chmod +x ${HOME}/.cargo/bin/subkey

gen-abi-bindings: install-deps ## Generates GO ABI Bindings
	npm install --prefix build/centrifuge-ethereum-contracts
	npm run compile --prefix build/centrifuge-ethereum-contracts
	@mkdir ./tmp
	@cat build/centrifuge-ethereum-contracts/build/contracts/Identity.json | jq '.abi' > ./tmp/id.abi
	@cat build/centrifuge-ethereum-contracts/build/contracts/IdentityFactory.json | jq '.abi' > ./tmp/idf.abi
	@abigen --abi ./tmp/id.abi --pkg ideth --type IdentityContract --out ${GOPATH}/src/github.com/centrifuge/go-centrifuge/identity/ideth/identity_contract.go
	@abigen --abi ./tmp/idf.abi --pkg ideth --type FactoryContract --out ${GOPATH}/src/github.com/centrifuge/go-centrifuge/identity/ideth/factory_contract.go
	@rm -Rf ./tmp

install: install-deps ## Builds and Install binary
	@go install -ldflags "-X github.com/centrifuge/go-centrifuge/version.gitCommit=`git rev-parse HEAD`" ./cmd/centrifuge/...

IMAGE_NAME?=centrifugeio/go-centrifuge
BRANCH=`git rev-parse --abbrev-ref HEAD`
GIT_SHORT_COMMIT=`git rev-parse --short HEAD`
TIMESTAMP=`date -u +%Y%m%d%H%M%S`
#TAG="${BRANCH}-${TIMESTAMP}-${GIT_SHORT_COMMIT}"
TAG="${TIMESTAMP}-${GIT_SHORT_COMMIT}"

build-docker: ## Build Docker Image
	@echo "Building Docker Image"
	@docker build -t ${IMAGE_NAME}:${TAG} .

push-to-docker: build-docker ## push docker image to registry
	@echo "Pushing Artifacts"
	@docker tag "${IMAGE_NAME}:${TAG}" "${IMAGE_NAME}:latest"
	@echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin
	@docker push ${IMAGE_NAME}:latest
	@docker push ${IMAGE_NAME}:${TAG}
