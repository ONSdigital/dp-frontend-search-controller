
#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-frontend-search-controller
  make test-component
popd