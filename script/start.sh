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

# Make and fill config directory
echo "Checking config directory..."
mkdir -p ~/.config/emp
mkdir -p ~/.config/emp/log
touch ~/.config/emp/known_nodes.dat
if [ ! -f ~/.config/emp/msg.conf ]; then
  cp "$DIR/msg.conf.example" ~/.config/emp/msg.conf
fi
rm -rf ~/.config/emp/client
cp -r "$DIR/../client" ~/.config/emp/

# Kill existing process
if [ -f ~/.config/emp/pid ];
then
  echo 'Killing existing process...'
  kill -15 `cat ~/.config/emp/pid`
  rm -f ~/.config/emp/pid
fi

# Get Dependencies
echo "Installing dependencies..."
go get golang.org/x/crypto/ripemd160
go get github.com/BurntSushi/toml
go get github.com/gorilla/rpc
go get github.com/mxk/go-sqlite/sqlite3

# Install and go!
echo "Building and running..."
if `go install emp`; then
	"$GOPATH/bin/emp" "$HOME/.config/emp/" > ~/.config/emp/log/log_`date +%s` &
	echo $! > ~/.config/emp/pid

	# Get Ports
	PORTS=$(sed -n 's/.*port *= *\([^ ]*.*\)/\1/p' < ~/.config/emp/msg.conf)
	IFS=' ' read -a array <<< $PORTS

	# Final Output
	echo "Started EMP client on local port ${array[0]}."
	echo "Access local client at: http://localhost:${array[1]}"

	export GOPATH=$TMPGOPATH
	exit 0

else echo "Build Failed, could not start client."
fi

export GOPATH=$TMPGOPATH
exit -1
