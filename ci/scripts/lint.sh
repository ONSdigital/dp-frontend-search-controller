#!/bin/bash -eux

pushd dp-frontend-search-controller
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0
  go install github.com/kevinburke/go-bindata/go-bindata
  make lint
popd
