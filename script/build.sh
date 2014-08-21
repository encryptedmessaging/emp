#!/bin/bash
: '
    Copyright 2014 JARST, LLC
    
    This file is part of EMP.

    EMP is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    EMP is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with Foobar.  If not, see <http://www.gnu.org/licenses/>.
'

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
