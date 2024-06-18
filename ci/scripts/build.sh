#!/bin/bash -eux

pushd dp-frontend-search-controller
  go install github.com/kevinburke/go-bindata/go-bindata
  make build
  cp build/dp-frontend-search-controller Dockerfile.concourse ../build
popd
