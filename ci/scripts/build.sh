#!/bin/bash -eux

pushd dp-frontend-search-controller
  make build
  cp build/dp-frontend-search-controller Dockerfile.concourse ../build
popd
