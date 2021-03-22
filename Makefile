CURRENT_DIRECTORY := $(shell pwd)
TESTS_TO_RUN := $(shell go list ./... | grep -v /integrationTests/ | grep -v mock)

build:
	go build ./...

build-cmd:
	(cd cmd && go build)

clean-test:
	go clean -testcache ./...

clean: clean-test
	go clean -cache ./...
	go clean ./...

test: clean-test
	go test ./...

test-coverage:
	@echo "Running unit tests"
	CURRENT_DIRECTORY=$(CURRENT_DIRECTORY) go test -cover -coverprofile=coverage.txt -covermode=atomic -v ${TESTS_TO_RUN}
