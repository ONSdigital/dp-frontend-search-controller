#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-frontend-search-controller
  go install github.com/kevinburke/go-bindata/go-bindata
  make test-component
popd
