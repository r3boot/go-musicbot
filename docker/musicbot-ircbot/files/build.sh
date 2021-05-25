#!/bin/sh

set -x

BUILD_DIR='/build'
GOPATH='/go'
export GOPATH
#GO111MODULE=auto
#export GO111MODULE
cd $GOPATH
cd "${GOPATH}/src/github.com/r3boot/go-musicbot"
make clean musicbot
