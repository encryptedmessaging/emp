#!/bin/bash

# Check for go
echo "Checking for go command..."
if ! which go > /dev/null; then
  echo "Go command not found, please install it."
  exit -1
fi

# Setup environment variables
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
GOPATH=$DIR/..

$DIR/stop.sh

echo "Cleaning built packages..."
rm -rf $GOPATH/bin $GOPATH/pkg

