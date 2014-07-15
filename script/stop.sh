#!/bin/bash

# Kill existing process
if [ -f ~/.config/emp/pid ];
then
  echo "Killing emp server (pid="`cat ~/.config/emp/pid`")"
  kill -2 `cat ~/.config/emp/pid`
  rm ~/.config/emp/pid
else
  echo "Server not running, execute start.sh"
fi

