#!/bin/bash

TRAVIS_BRANCH?=`git rev-parse --abbrev-ref HEAD`
GIT_COMMIT=`git rev-parse HEAD`
GIT_SHORT_COMMIT=`git rev-parse --short HEAD`
TIMESTAMP=`date -u +%Y%m%d%H%M%S`
TAG="${TRAVIS_BRANCH}-${TIMESTAMP}-${GIT_SHORT_COMMIT}"
IMAGE_NAME?=centrifugeio/go-centrifuge
LD_FLAGS?="-X github.com/centrifuge/go-centrifuge/version.gitCommit=${GIT_COMMIT}"
GCLOUD_SERVICE?="./build/peak-vista-185616-9f70002df7eb.json"

# Default TAGINSTANCE for standalone targets
TAGINSTANCE="${TAG}"

# GOBIN needs to be set to ensure govendor can actually be found and executed
PATH=$(shell printenv PATH):$(GOBIN)

# Lock metalinter version
GOMETALINTER_VERSION="v3.0.0"

.PHONY: help vendor

help: ## Show this help message.
	@echo 'usage: make [target] ...'
	@echo
	@echo 'targets:'
	@egrep '^(.+)\:\ ##\ (.+)' ${MAKEFILE_LIST} | column -t -c 2 -s ':#'

clean: ##clean vendor's folder. Should be run before a make install
	@echo 'cleaning previous /vendor folder'
	@rm -rf vendor/
	@echo 'done cleaning'

install-deps: ## Install Dependencies
	@go get github.com/goware/modvendor@v0.3.0
	@go get github.com/jteeuwen/go-bindata@v3.0.7+incompatible
	@go get github.com/swaggo/swag/cmd/swag@v1.6.7
	@go get github.com/ethereum/go-ethereum/cmd/abigen@v1.9.15
	@go get github.com/karalabe/xgo@v0.0.0-20191115072854-c5ccff8648a7
	@git submodule update --init --recursive
	@curl -L https://git.io/vp6lP | sh -s ${GOMETALINTER_VERSION}
	@mv ./bin/* $(GOPATH)/bin/; rm -rf ./bin

vendor: install-deps ## Create the Go vendor folder with dependencies
	@go mod tidy
	@go mod vendor
	@modvendor -copy="**/*.c **/*.h"

lint-check: ## runs linters on go code
	@gometalinter --exclude=anchors/service.go  --disable-all --enable=golint --enable=goimports --enable=vet --enable=nakedret \
	--vendor --skip=resources --skip=testingutils --deadline=1m ./...;

format-go: ## formats go code
	@goimports -w .

gen-swagger: ## generates the swagger documentation
	swag init -g ./httpapi/router.go -o ./httpapi
	rm -rf ./httpapi/docs.go ./httpapi/swagger.yaml

generate: ## autogenerate go files for config
	go generate -mod=readonly ./config/configuration.go

install-subkey: ## installs subkey
	curl https://getsubstrate.io -sSf | bash -s -- --fast
	cargo install --force --git https://github.com/paritytech/substrate subkey

gen-abi-bindings: vendor ## Generates GO ABI Bindings
	npm install --prefix build/centrifuge-ethereum-contracts
	npm run compile --prefix build/centrifuge-ethereum-contracts
	@mkdir ./tmp
	@cat build/centrifuge-ethereum-contracts/build/contracts/Identity.json | jq '.abi' > ./tmp/id.abi
	@cat build/centrifuge-ethereum-contracts/build/contracts/IdentityFactory.json | jq '.abi' > ./tmp/idf.abi
	@abigen --abi ./tmp/id.abi --pkg ideth --type IdentityContract --out ${GOPATH}/src/github.com/centrifuge/go-centrifuge/identity/ideth/identity_contract.go
	@abigen --abi ./tmp/idf.abi --pkg ideth --type FactoryContract --out ${GOPATH}/src/github.com/centrifuge/go-centrifuge/identity/ideth/factory_contract.go
	@rm -Rf ./tmp

install: vendor ## Builds and Install binary for development
	@go install -mod=readonly ./cmd/centrifuge/...

build-darwin-amd64: vendor ## Build darwin/amd64
	@echo "Building darwin-10.10-amd64 with flags [${LD_FLAGS}]"
	@mkdir -p build/darwin-amd64
	@xgo -go 1.14.x -dest build/darwin-amd64 -targets=darwin-10.10/amd64 -ldflags=${LD_FLAGS} ./cmd/centrifuge/
	@mv build/darwin-amd64/centrifuge-darwin-10.10-amd64 build/darwin-amd64/centrifuge
	$(eval TAGINSTANCE := $(shell echo ${TAG}))
	@tar -zcvf cent-api-darwin-10.10-amd64-${TAGINSTANCE}.tar.gz -C build/darwin-amd64/ .

build-linux-amd64: vendor ## Build linux/amd64
	@echo "Building amd64 with flags [${LD_FLAGS}]"
	@mkdir -p build/linux-amd64
	@xgo -go 1.14.x -dest build/linux-amd64 -targets=linux/amd64 -ldflags=${LD_FLAGS} ./cmd/centrifuge/
	@mv build/linux-amd64/centrifuge-linux-amd64 build/linux-amd64/centrifuge
	$(eval TAGINSTANCE := $(shell echo ${TAG}))
	@tar -zcvf cent-api-linux-amd64-${TAGINSTANCE}.tar.gz -C build/linux-amd64/ .

build-docker: ## Build Docker Image
	@echo "Building Docker Image"
	@docker build -t ${IMAGE_NAME}:${TAGINSTANCE} .

build-ci: build-linux-amd64 build-docker ## Builds + Push all artifacts
	@echo "Building/Pushing Artifacts for CI"
	@gcloud auth activate-service-account --key-file ${GCLOUD_SERVICE}
	@gsutil cp cent-api-*-${TAGINSTANCE}.tar.gz gs://centrifuge-artifact-releases/
	@gsutil acl ch -u AllUsers:R gs://centrifuge-artifact-releases/cent-api-*-${TAGINSTANCE}.tar.gz
	@echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin
	@docker tag "${IMAGE_NAME}:${TAGINSTANCE}" "${IMAGE_NAME}:latest"
	@docker push ${IMAGE_NAME}:latest
	@docker push ${IMAGE_NAME}:${TAGINSTANCE}
