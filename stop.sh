#!/bin/bash

# Kill existing process
if [ -f ./pid ];
then
  kill -9 `cat ./pid`
  rm pid
fi
