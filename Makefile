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

semi-integration-tests: clean-test
	@docker compose -f docker/docker-compose.yml build
	@docker compose -f docker/docker-compose.yml up & go test ./integrationTests/relayers/slowTests/... -v -timeout 20m -tags integration
	@docker compose -f docker/docker-compose.yml down -v

slow-tests-01: clean-test
	@docker compose -f docker/docker-compose.yml build
	@docker compose -f docker/docker-compose.yml up & go test ./integrationTests/relayers/slowTests/01happyFlowTests/... -v -timeout 20m -tags slow
	@docker compose -f docker/docker-compose.yml down -v

slow-tests-02: clean-test
	@docker compose -f docker/docker-compose.yml build
	@docker compose -f docker/docker-compose.yml up & go test ./integrationTests/relayers/slowTests/02setupErrorTests/... -v -timeout 20m -tags slow
	@docker compose -f docker/docker-compose.yml down -v

slow-tests-03: clean-test
	@docker compose -f docker/docker-compose.yml build
	@docker compose -f docker/docker-compose.yml up & go test ./integrationTests/relayers/slowTests/03edgeCaseTests/... -v -timeout 20m -tags slow
	@docker compose -f docker/docker-compose.yml down -v

slow-tests-04: clean-test
	@docker compose -f docker/docker-compose.yml build
	@docker compose -f docker/docker-compose.yml up & go test ./integrationTests/relayers/slowTests/04refundTestsWithMalformedSCData/... -v -timeout 20m -tags slow
	@docker compose -f docker/docker-compose.yml down -v

slow-tests-05: clean-test
	@docker compose -f docker/docker-compose.yml build
	@docker compose -f docker/docker-compose.yml up & go test ./integrationTests/relayers/slowTests/05refundTestsWrongFunction/... -v -timeout 20m -tags slow
	@docker compose -f docker/docker-compose.yml down -v

slow-tests-06: clean-test
	@docker compose -f docker/docker-compose.yml build
	@docker compose -f docker/docker-compose.yml up & go test ./integrationTests/relayers/slowTests/06refundTestsWrongGasLimit/... -v -timeout 20m -tags slow
	@docker compose -f docker/docker-compose.yml down -v

slow-tests-07: clean-test
	@docker compose -f docker/docker-compose.yml build
	@docker compose -f docker/docker-compose.yml up & go test ./integrationTests/relayers/slowTests/07refundTestsWrongParams/... -v -timeout 20m -tags slow
	@docker compose -f docker/docker-compose.yml down -v

slow-tests-08: clean-test
	@docker compose -f docker/docker-compose.yml build
	@docker compose -f docker/docker-compose.yml up & go test ./integrationTests/relayers/slowTests/08refundTestsOther/... -v -timeout 20m -tags slow
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

cli-docs:
	cd ./cmd/scCallsExecutor && go build
	cd ./cmd && bash ./CLI.md.sh

check-cli-md: cli-docs
	@status=$$(git status --porcelain | grep CLI); \
    	if [ ! -z "$${status}" ]; \
    	then \
    		echo "Error - please update all CLI.md files by running the 'cli-docs' or 'check-cli-md' from Makefile!"; \
    		exit 1; \
    	fi
