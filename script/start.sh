#!/bin/bash

# Kill existing process
if [ -f ./pid ];
then
  echo 'killing existing process...'
  kill -9 `cat ./pid`
  rm pid
fi

GOPATH=`pwd`:$GOPATH go build EMP.go
./EMP > log_`date +%s` &
echo $! >> ./pid
sleep 1
firefox http://localhost:8080/
