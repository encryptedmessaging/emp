#!/bin/zsh

# Kill existing process
if [ -f ./pid ];
then
  echo 'killing existing process...'
  kill -9 `cat ./pid`
fi


GOPATH=`pwd`:$GOPATH go run EMP.go >> log &
echo $! > ./pid
sleep 1
firefox http://localhost:8080/
