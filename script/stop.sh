#!/bin/bash

# Kill existing process
if [ -f ~/.config/emp/pid ];
then
  kill -2 `cat ~/.config/emp/pid`
  rm ~/.config/emp/pid
fi
