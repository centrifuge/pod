GIT_COMMIT=`git rev-parse HEAD`
GIT_SHORT_COMMIT=`git rev-parse --short HEAD`
TIMESTAMP=`date -u +%Y%m%d`
TAG="${TIMESTAMP}-${GIT_SHORT_COMMIT}"
IMAGE_NAME?=centrifugeio/go-centrifuge
LD_FLAGS?="-X github.com/CentrifugeInc/go-centrifuge/centrifuge/version.gitCommit=${GIT_COMMIT}"
GCLOUD_SERVICE?="peak-vista-185616-9f70002df7eb.json"

.PHONY: help

help: ## Show this help message.
	@echo 'usage: make [target] ...'
	@echo
	@echo 'targets:'
	@egrep '^(.+)\:\ ##\ (.+)' ${MAKEFILE_LIST} | column -t -c 2 -s ':#'

install-deps: ## Install Dependencies
	@command -v dep >/dev/null 2>&1 || go get -u github.com/golang/dep/...
	@dep ensure

install: ## Builds and Install binary for development
install: install-deps
	@go install ./centrifuge/

install-xgo: ## Install XGO
	@echo "Ensuring XGO is installed"
	@command -v xgo >/dev/null 2>&1 || go get github.com/karalabe/xgo

build-linux-amd64: ## Build linux/amd64
build-linux-amd64: install-xgo
	@echo "Building amd64 with flags [${LD_FLAGS}]"
	@xgo -dest build/linux-amd64 -targets=linux/amd64 -ldflags=${LD_FLAGS} ./centrifuge
	@mv build/linux-amd64/centrifuge-linux-amd64 build/linux-amd64/centrifuge
	@tar -zcvf cent-api-linux-amd64-${TAG}.tar.gz -C build/linux-amd64/ ./centrifuge

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
