CURRENT_DIRECTORY := $(shell pwd)
TESTS_TO_RUN := $(shell go list ./... | grep -v /integrationTests/ | grep -v mock)

build:
	go build ./...

build-cmd:
	(cd cmd && go build)

clean-test:
	go clean -testcache

test: clean-test
	go test ./...

test-coverage:
	@echo "Running unit tests"
	CURRENT_DIRECTORY=$(CURRENT_DIRECTORY) go test -cover -coverprofile=coverage.txt -covermode=atomic -v ${TESTS_TO_RUN}

slow-tests: clean-test
	@docker compose -f docker/docker-compose.yml build
	@docker compose -f docker/docker-compose.yml up & go test ./integrationTests/... -v -timeout 60m -tags slow
	@docker compose -f docker/docker-compose.yml down -v

lint-install:
ifeq (,$(wildcard test -f bin/golangci-lint))
	@echo "Installing golint"
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s
endif

run-lint:
	@echo "Running golint"
	bin/golangci-lint run --max-issues-per-linter 0 --max-same-issues 0 --timeout=2m

lint: lint-install run-lint