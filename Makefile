#!/bin/bash
.PHONY: help

help: ## Show this help message.
	@echo 'usage: make [target] ...'
	@echo
	@echo 'targets:'
	@egrep '^(.+)\:\ ##\ (.+)' ${MAKEFILE_LIST} | column -t -c 2 -s ':#'

install-deps: ## Install Dependencies
	@go mod tidy
	@go install github.com/jteeuwen/go-bindata/go-bindata
	@go install github.com/swaggo/swag/cmd/swag
	@git submodule update --init --recursive
	@command -v golangci-lint >/dev/null 2>&1 || (curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${GOPATH}/bin v1.45.2)


lint-check: ## runs linters on go code
	@golangci-lint run --skip-dirs=build/*  --disable-all --enable=revive --enable=goimports --enable=vet --enable=nakedret \
	--enable=unused --skip-dirs=resources --skip-dirs=testingutils --timeout=2m ./...;

gen-swagger: ## generates the swagger documentation
	swag init --parseDependency -g ./http/router.go -o ./http
	rm -rf ./http/docs.go ./http/swagger.yaml

generate: ## autogenerate go files for config
	go generate ./config/configuration.go

run-unit-tests:
	@rm -rf profile.out
	go test ./... -v -race -coverprofile=profile.out -covermode=atomic -tags=unit

run-integration-tests:
	@rm -rf profile.out
	go test ./... -v -race -coverprofile=profile.out -covermode=atomic -tags=integration -timeout 30m

run-testworld-tests:
	@rm -rf profile.out
	go test ./... -v -race -coverprofile=profile.out -covermode=atomic -tags=testworld -timeout 60m

install: ## Builds and Install binary
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
	@echo "${DOCKERHUB_TOKEN}" | docker login -u "${DOCKERHUB_USERNAME}" --password-stdin
	@#docker push ${IMAGE_NAME}:latest
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

generate-mocks:
	@docker run -v `pwd`:/app -w /app --entrypoint /bin/sh vektra/mockery:v2.13.0-beta.1 -c 'go generate ./...'

targetDir?=${HOME}/centrifuge/testing
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
		--centchainsecret="//Alice" --centchainurl="ws://localhost:9946" --network=testing; \
	fi
	@echo "Starting centrifuge node..."
	@CENT_ETHEREUM_ACCOUNTS_MAIN_KEY='${ethAccountKey}' CENT_ETHEREUM_ACCOUNTS_MAIN_PASSWORD="" ./centrifuge run -c "${targetDir}"/config.yaml
