#!/bin/sh
export GOPATH="`pwd`"
mkdir -p ./src/
cd ./src/

test -d ./sample || ln -s ../sample/ ./sample
test -d ./termwindow || ln -s ../termwindow/ ./termwindow

cd ..
ls -1 src/sample | while read row ; do
  GOOS=linux GOARCH=amd64 go install -ldflags "-s -w" sample/$row
#  GOOS=windows GOARCH=amd64 go install -ldflags "-s -w" sample/$row
#  GOOS=darwin GOARCH=amd64 go install -ldflags "-s -w" sample/$row
done
