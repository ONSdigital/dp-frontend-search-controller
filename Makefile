BINPATH ?= build

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)

SERVICE_PATH = github.com/ONSdigital/dp-frontend-search-controller/service

LDFLAGS = -ldflags "-X $(SERVICE_PATH).BuildTime=$(BUILD_TIME) -X $(SERVICE_PATH).GitCommit=$(GIT_COMMIT) -X $(SERVICE_PATH).Version=$(VERSION)"

.PHONY: all
all: audit test build

.PHONY: audit
audit:
	go list -m all | nancy sleuth

.PHONY: build
build:
	go build -tags 'production' $(LDFLAGS) -o $(BINPATH)/dp-frontend-search-controller

.PHONY: debug
debug:
	go build -tags 'debug' $(LDFLAGS) -o $(BINPATH)/dp-frontend-search-controller
	HUMAN_LOG=1 DEBUG=1 $(BINPATH)/dp-frontend-search-controller

.PHONY: test
test:
	go test -race -cover ./...

.PHONY:	test-component
test-component:
	go test -race -cover -coverprofile="coverage.txt" -coverpkg=github.com/ONSdigital/dp-frontend-search-controller/... -component

.PHONY: convey
convey:
	goconvey ./...