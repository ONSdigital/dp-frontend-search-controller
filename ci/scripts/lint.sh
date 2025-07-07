#!/bin/bash -eux

pushd dp-frontend-search-controller
  go install github.com/kevinburke/go-bindata/go-bindata
  make lint
popd
