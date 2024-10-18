#!/bin/bash -eux

pushd dp-frontend-search-controller
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0
  make lint
popd
