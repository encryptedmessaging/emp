#!/bin/bash

TMPGOPATH=$GOPATH

# Check for go
echo "Checking for go command..."
if ! which go > /dev/null; then
  echo "Go command not found, please install it."
  exit -1
fi

# Setup environment variables
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export GOPATH=$DIR/..

# Get Dependencies
echo "Installing dependencies..."
go get code.google.com/p/go.crypto/ripemd160
go get github.com/BurntSushi/toml
go get github.com/gorilla/rpc
go get github.com/mxk/go-sqlite/sqlite3

# Install and go!
echo "Building..."
if `go install emp`; then
  echo "Build succeeded."
  exit 0
else echo "Build Failed, could not start client."
fi

export GOPATH=$TMPGOPATH
exit -1
