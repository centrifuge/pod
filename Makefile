#!/bin/bash

TRAVIS_BRANCH?=`git rev-parse --abbrev-ref HEAD`
GIT_COMMIT=`git rev-parse HEAD`
GIT_SHORT_COMMIT=`git rev-parse --short HEAD`
TIMESTAMP=`date -u +%Y%m%d%H%M%S`
TAG="${TRAVIS_BRANCH}-${TIMESTAMP}-${GIT_SHORT_COMMIT}"
IMAGE_NAME?=centrifugeio/go-centrifuge
LD_FLAGS?="-X github.com/centrifuge/go-centrifuge/version.gitCommit=${GIT_COMMIT}"
GCLOUD_SERVICE?="./build/peak-vista-185616-9f70002df7eb.json"
GO_REQUIRED_VERSION=11
GO_INSTALLED_VERSION=`go version | awk '{split($$3, a, "."); print a[2]}'`

# Default TAGINSTANCE for standalone targets
TAGINSTANCE="${TAG}"

# GOBIN needs to be set to ensure govendor can actually be found and executed
PATH=$(shell printenv PATH):$(GOBIN)

# Lock metalinter version
GOMETALINTER_VERSION="v2.0.12"

.PHONY: help

help: ## Show this help message.
	@echo 'usage: make [target] ...'
	@echo
	@echo 'targets:'
	@egrep '^(.+)\:\ ##\ (.+)' ${MAKEFILE_LIST} | column -t -c 2 -s ':#'

check-version: ## checks if the compatible go version is installed
	@if [ ${GO_INSTALLED_VERSION} -gt ${GO_REQUIRED_VERSION} ]; then\
	 echo "please install GO 1.${GO_REQUIRED_VERSION}.xx to proceed with installation.";\
	 exit 1;\
	fi

clean: ##clean vendor's folder. Should be run before a make install
	@echo 'cleaning previous /vendor folder'
	@rm -rf vendor/
	@echo 'done cleaning'

install-deps: ## Install Dependencies
	@command -v dep >/dev/null 2>&1 || go get -u github.com/golang/dep/...
	@dep ensure
	@curl -L https://git.io/vp6lP | sh -s ${GOMETALINTER_VERSION}
	@git submodule update --init --recursive
	@mv ./bin/* $(GOPATH)/bin/; rm -rf ./bin

lint-check: ## runs linters on go code
	@gometalinter --exclude=anchors/service.go  --disable-all --enable=golint --enable=goimports --enable=vet --enable=nakedret \
	--enable=staticcheck --vendor --skip=resources --skip=testingutils --deadline=1m ./...;

format-go: ## formats go code
	@goimports -w .

gen-swagger: ## generates the swagger documentation
	swag init -g ./httpapi/router.go -o ./httpapi
	rm -rf ./httpapi/docs.go ./httpapi/swagger.yaml

generate: ## autogenerate go files for config
	go generate ./config/configuration.go

create-testworld-config:
	@if [ "${TESTWORLD_NETWORK}" != "" ]; then\
		echo ${TESTWORLD_CONFIG} | base64 -D > ./testworld/configs/${TESTWORLD_NETWORK}.json;\
		echo ${TESTWORLD_ETH_KEY} | base64 -D > ./testworld/configs/${TESTWORLD_NETWORK}.eth;\
	fi

install-subkey: ## installs subkey
	curl https://getsubstrate.io -sSf | bash -s -- --fast
	cargo install --force --git https://github.com/paritytech/substrate subkey

vendorinstall: ## Installs all protobuf dependencies with go-vendorinstall
	go install github.com/centrifuge/go-centrifuge/vendor/github.com/roboll/go-vendorinstall
	go-vendorinstall golang.org/x/tools/cmd/goimports
	go-vendorinstall github.com/swaggo/swag/cmd/swag
	go get -u github.com/jteeuwen/go-bindata/...

abigen-install: ## Installs ABIGEN from vendor
abigen-install: vendorinstall
	go-vendorinstall github.com/ethereum/go-ethereum/cmd/abigen

gen-abi-bindings: ## Generates GO ABI Bindings
gen-abi-bindings: install-deps abigen-install
	npm install --prefix vendor/github.com/centrifuge/centrifuge-ethereum-contracts
	npm run compile --prefix vendor/github.com/centrifuge/centrifuge-ethereum-contracts
	@cat vendor/github.com/centrifuge/centrifuge-ethereum-contracts/build/contracts/Identity.json | jq '.abi' > tmp/contracts/id.abi
	@cat vendor/github.com/centrifuge/centrifuge-ethereum-contracts/build/contracts/IdentityFactory.json | jq '.abi' > tmp/contracts/idf.abi
	@abigen --abi tmp/contracts/id.abi --pkg ideth --type IdentityContract --out ${GOPATH}/src/github.com/centrifuge/go-centrifuge/identity/ideth/identity_contract.go
	@abigen --abi tmp/contracts/idf.abi --pkg ideth --type FactoryContract --out ${GOPATH}/src/github.com/centrifuge/go-centrifuge/identity/ideth/factory_contract.go
	@rm -Rf ./tmp

install: ## Builds and Install binary for development
install: check-version install-deps vendorinstall
	@go install ./cmd/centrifuge/...

install-xgo: ## Install XGO
	@echo "Ensuring XGO is installed"
	@command -v xgo >/dev/null 2>&1 || go get github.com/karalabe/xgo

build-darwin-amd64: ## Build darwin/amd64
build-darwin-amd64: install-xgo
	@echo "Building darwin-10.10-amd64 with flags [${LD_FLAGS}]"
	@mkdir -p build/darwin-amd64
	@xgo -go 1.11.x -dest build/darwin-amd64 -targets=darwin-10.10/amd64 -ldflags=${LD_FLAGS} ./cmd/centrifuge/
	@mv build/darwin-amd64/centrifuge-darwin-10.10-amd64 build/darwin-amd64/centrifuge
	$(eval TAGINSTANCE := $(shell echo ${TAG}))
	@tar -zcvf cent-api-darwin-10.10-amd64-${TAGINSTANCE}.tar.gz -C build/darwin-amd64/ .

build-linux-amd64: ## Build linux/amd64
build-linux-amd64: install-xgo
	@echo "Building amd64 with flags [${LD_FLAGS}]"
	@mkdir -p build/linux-amd64
	@xgo -go 1.11.x -dest build/linux-amd64 -targets=linux/amd64 -ldflags=${LD_FLAGS} ./cmd/centrifuge/
	@mv build/linux-amd64/centrifuge-linux-amd64 build/linux-amd64/centrifuge
	$(eval TAGINSTANCE := $(shell echo ${TAG}))
	@tar -zcvf cent-api-linux-amd64-${TAGINSTANCE}.tar.gz -C build/linux-amd64/ .

build-docker: ## Build Docker Image
build-docker:
	@echo "Building Docker Image"
	@docker build -t ${IMAGE_NAME}:${TAGINSTANCE} .

build-ci: ## Builds + Push all artifacts
build-ci: build-linux-amd64 build-docker
	@echo "Building/Pushing Artifacts for CI"
	@gcloud auth activate-service-account --key-file ${GCLOUD_SERVICE}
	@gsutil cp cent-api-*-${TAGINSTANCE}.tar.gz gs://centrifuge-artifact-releases/
	@gsutil acl ch -u AllUsers:R gs://centrifuge-artifact-releases/cent-api-*-${TAGINSTANCE}.tar.gz
	@echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin
	@docker tag "${IMAGE_NAME}:${TAGINSTANCE}" "${IMAGE_NAME}:latest"
	@docker push ${IMAGE_NAME}:latest
	@docker push ${IMAGE_NAME}:${TAGINSTANCE}
