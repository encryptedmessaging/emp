#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Make config directory
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
  echo 'killing existing process...'
  kill -15 `cat ~/.config/emp/pid`
  rm -f ~/.config/emp/pid
fi

if `go install emp`; then
	$GOPATH/bin/emp > ~/.config/emp/log/log_`date +%s` &
	echo $! > ~/.config/emp/pid
else echo "Build Failed, could not start client."
fi
