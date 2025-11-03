BINPATH ?= build

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)
LOCAL_RENDERER_IN_USE = $(shell grep -c "github.com/ONSdigital/dis-design-system-go =" go.mod)

SERVICE_PATH = github.com/ONSdigital/dp-frontend-search-controller/service

LDFLAGS_DEBUG = -ldflags "-X $(SERVICE_PATH).BuildTime=$(BUILD_TIME) -X $(SERVICE_PATH).GitCommit=$(GIT_COMMIT) -X $(SERVICE_PATH).Version=$(VERSION)"
LDFLAGS_BUILD = -ldflags "-X $(SERVICE_PATH).BuildTime=$(BUILD_TIME) -X $(SERVICE_PATH).GitCommit=$(GIT_COMMIT) -X $(SERVICE_PATH).Version=$(VERSION) -X github.com/ONSdigital/dp-frontend-search-controller/config.RendererVersion=$(APP_RENDERER_VERSION)"

.PHONY: all
all: audit test build

.PHONY: audit
audit: generate-prod
	dis-vulncheck --build-tags=production

.PHONY: build
build: generate-prod
	go build -tags 'production' $(LDFLAGS_BUILD) -o $(BINPATH)/dp-frontend-search-controller

.PHONY: debug
debug: generate-debug
	go build -tags 'debug' $(LDFLAGS_DEBUG) -o $(BINPATH)/dp-frontend-search-controller
	HUMAN_LOG=1 DEBUG=1 $(BINPATH)/dp-frontend-search-controller

.PHONY: debug-watch
debug-watch: 
	reflex -d none -c ./reflex

.PHONY: lint
lint: ## Used in ci to run linters against Go code
	cp assets/assets.go assets/assets.go.bak
	echo 'func Asset(_ string) ([]byte, error) { return nil, nil }' >> assets/assets.go
	echo 'func AssetNames() []string { return []string{} }' >> assets/assets.go
	gofmt -w assets/assets.go
	golangci-lint run ./... || { echo "Linting failed, restoring original assets.go"; mv assets/assets.go.bak assets/assets.go; exit 1; }
	mv assets/assets.go.bak assets/assets.go

.PHONY: lint-local
lint-local: ## Use locally to run linters against Go code
	golangci-lint run ./...

.PHONY: test
test: generate-prod
	go test -race -cover -tags 'production' ./...

.PHONY:	test-component
test-component: generate-prod
	go test -race -cover -tags 'production' -coverprofile="coverage.txt" -coverpkg=github.com/ONSdigital/dp-frontend-search-controller/... -component

.PHONY: convey
convey:
	goconvey ./...

.PHONY: fetch-renderer
fetch-renderer-lib:
ifeq ($(LOCAL_RENDERER_IN_USE), 1)
 $(info local renderer in use)
 $(eval CORE_ASSETS_PATH = $(shell grep -w "github.com/ONSdigital/dis-design-system-go =>" go.mod | awk -F '=> ' '{print $$2}' | tr -d '"'))
else
 $(eval APP_RENDERER_VERSION=$(shell grep "github.com/ONSdigital/dis-design-system-go" go.mod | cut -d ' ' -f2 ))
 $(eval CORE_ASSETS_PATH = $(shell go get github.com/ONSdigital/dis-design-system-go@$(APP_RENDERER_VERSION) && go list -f '{{.Dir}}' -m github.com/ONSdigital/dis-design-system-go))
endif

.PHONY: generate-debug
generate-debug: fetch-renderer
	cd assets; go run github.com/kevinburke/go-bindata/go-bindata -prefix $(CORE_ASSETS_PATH)/assets -debug -o data.go -pkg assets locales/... templates/... $(CORE_ASSETS_PATH)/assets/locales/... $(CORE_ASSETS_PATH)/assets/templates/...
	{ printf "//go:build debug\n"; cat assets/data.go; } > assets/debug.go.new
	mv assets/debug.go.new assets/data.go

.PHONY: generate-prod
generate-prod: fetch-renderer
	cd assets; go run github.com/kevinburke/go-bindata/go-bindata -prefix $(CORE_ASSETS_PATH)/assets -o data.go -pkg assets locales/... templates/... $(CORE_ASSETS_PATH)/assets/locales/... $(CORE_ASSETS_PATH)/assets/templates/...
	{ printf "//go:build production\n"; cat assets/data.go; } > assets/data.go.new
	mv assets/data.go.new assets/data.go
