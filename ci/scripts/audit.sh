#!/bin/bash -eux

export cwd=$(pwd)

pushd $cwd/dp-frontend-search-controller
  make audit
popd
