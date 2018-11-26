TRAVIS_BRANCH?=`git rev-parse --abbrev-ref HEAD`
GIT_COMMIT=`git rev-parse HEAD`
GIT_SHORT_COMMIT=`git rev-parse --short HEAD`
TIMESTAMP=`date -u +%Y%m%d%H`
TAG="${TRAVIS_BRANCH}-${TIMESTAMP}-${GIT_SHORT_COMMIT}"
IMAGE_NAME?=centrifugeio/go-centrifuge
LD_FLAGS?="-X github.com/centrifuge/go-centrifuge/version.gitCommit=${GIT_COMMIT}"
GCLOUD_SERVICE?="./build/peak-vista-185616-9f70002df7eb.json"

# GOBIN needs to be set to ensure govendor can actually be found and executed 
PATH=$(shell printenv PATH):$(GOBIN)

# If you need to overwrite PROTOTOOL_BIN, you can set this environment variable.
PROTOTOOL_BIN ?=$(shell which prototool)


.PHONY: help

help: ## Show this help message.
	@echo 'usage: make [target] ...'
	@echo
	@echo 'targets:'
	@egrep '^(.+)\:\ ##\ (.+)' ${MAKEFILE_LIST} | column -t -c 2 -s ':#'

install-deps: ## Install Dependencies
	@command -v dep >/dev/null 2>&1 || go get -u github.com/golang/dep/...
	@dep ensure
	@npm --prefix ./build  install
	@curl -L https://git.io/vp6lP | sh
	@mv ./bin/* $(GOPATH)/bin/; rm -rf ./bin

lint-check: ## formats go code
	@gometalinter --disable-all --enable=golint --enable=goimports --enable=vet --enable=nakedret \
	--enable=staticcheck --vendor --skip=resources --skip=testingutils --skip=protobufs  --deadline=1m ./...;

format-go: ## formats go code
	@goimports -w .

proto-lint: ## runs prototool lint
	$(PROTOTOOL_BIN) lint

proto-gen-go: ## generates the go bindings
	$(PROTOTOOL_BIN) gen

proto-all: ## runs prototool all
	$(PROTOTOOL_BIN) all

gen-swagger: ## generates the swagger documentation
	npm --prefix ./build  run build_swagger

generate: ## autogenerate go files for config
	go generate ./config/configuration.go

vendorinstall: ## Installs all protobuf dependencies with go-vendorinstall
	go install github.com/centrifuge/go-centrifuge/vendor/github.com/roboll/go-vendorinstall
	go-vendorinstall github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	go-vendorinstall github.com/golang/protobuf/protoc-gen-go
	go-vendorinstall github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
	go get -u github.com/jteeuwen/go-bindata/...

install: ## Builds and Install binary for development
install: install-deps vendorinstall
	@go install ./...

install-xgo: ## Install XGO
	@echo "Ensuring XGO is installed"
	@command -v xgo >/dev/null 2>&1 || go get github.com/karalabe/xgo

build-linux-amd64: ## Build linux/amd64
build-linux-amd64: install-xgo
	@echo "Building amd64 with flags [${LD_FLAGS}]"
	@mkdir -p build/linux-amd64
	@xgo -go 1.11.x -dest build/linux-amd64 -targets=linux/amd64 -ldflags=${LD_FLAGS} .
	@mv build/linux-amd64/go-centrifuge-linux-amd64 build/linux-amd64/go-centrifuge
	@tar -zcvf cent-api-linux-amd64-${TAG}.tar.gz -C build/linux-amd64/ .

build-docker: ## Build Docker Image
build-docker:
	@echo "Building Docker Image"
	@docker build -t ${IMAGE_NAME}:${TAG} .

build-ci: ## Builds + Push all artifacts
build-ci: build-linux-amd64 build-docker
	@echo "Building/Pushing Artifacts for CI"
	@gcloud auth activate-service-account --key-file ${GCLOUD_SERVICE}
	@gsutil cp cent-api-*-${TAG}.tar.gz gs://centrifuge-artifact-releases/
	@gsutil acl ch -u AllUsers:R gs://centrifuge-artifact-releases/cent-api-*-${TAG}.tar.gz
	@echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin
	@docker tag "${IMAGE_NAME}:${TAG}" "${IMAGE_NAME}:latest"
	@docker push ${IMAGE_NAME}:latest
	@docker push ${IMAGE_NAME}:${TAG}
