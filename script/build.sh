#!/bin/bash

TMPGOPATH=$GOPATH

# Check for go
echo "Checking for go command..."
if ! which go > /dev/null; then
  echo "Go command not found, please install it."
  exit -1
fi

if ! which arm-linux-gnueabi-gcc > /dev/null; then
  echo "Build Script requires arm-linux-gnueabi-gcc compiler. Change line 12 of script/build.sh for a different compiler."
  exit -1
fi

# Setup environment variables
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export GOPATH=$DIR/..

# Get Dependencies
echo "Installing dependencies..."
CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOARCH=arm GOOS=linux GOARM=6 go get code.google.com/p/go.crypto/ripemd160
CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOARCH=arm GOOS=linux GOARM=6 go get github.com/BurntSushi/toml
CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOARCH=arm GOOS=linux GOARM=6 go get github.com/gorilla/rpc
CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOARCH=arm GOOS=linux GOARM=6 go get github.com/mxk/go-sqlite/sqlite3

# Install and go!
echo "Building..."
if `CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm go build --ldflags '-extld "arm-linux-gnueabi-gcc" -extldflags "-static"' emp`; then
  echo "Build succeeded."
  exit 0
else echo "Build Failed, could not start client."
fi

export GOPATH=$TMPGOPATH
exit -1
