#!/bin/bash

# Kill existing process
if [ -f ./pid ];
then
  kill -2 `cat ./pid`
  rm pid
fi
