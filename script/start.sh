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

# Make and fill config directory
echo "Checking config directory..."
mkdir -p ~/.config/emp
mkdir -p ~/.config/emp/log
if [ ! -f ~/.config/emp/msg.conf ]; then
  cp $DIR/msg.conf.example ~/.config/emp/msg.conf
fi
rm -rf ~/.config/emp/client
cp -r $DIR/../client ~/.config/emp/

# Kill existing process
if [ -f ~/.config/emp/pid ];
then
  echo 'Killing existing process...'
  kill -15 `cat ~/.config/emp/pid`
  rm -f ~/.config/emp/pid
fi

# Get Dependencies
echo "Installing dependencies..."
go get code.google.com/p/go.crypto/ripemd160
go get github.com/BurntSushi/toml
go get github.com/gorilla/rpc
go get github.com/mxk/go-sqlite/sqlite3

# Install and go!
echo "Building and running..."
if `go install emp`; then
	$GOPATH/bin/emp > ~/.config/emp/log/log_`date +%s` &
	echo $! > ~/.config/emp/pid
else echo "Build Failed, could not start client."
fi

# Get Ports
PORTS=$(sed -n 's/.*port *= *\([^ ]*.*\)/\1/p' < ~/.config/emp/msg.conf)
IFS=' ' read -a array <<< $PORTS

# Final Output
echo "Started EMP client on local port ${array[0]}."
echo "Access local client at: http://localhost:${array[1]}"
