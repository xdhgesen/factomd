#!/usr/bin/env bash

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)" # get dir containing this script
cd $DIR                                                             # always from from script dirUUU

# create a static shared obj
go build -v -o libfactomd.so \
    -ldflags "-X github.com/FactomProject/factomd/engine.Build=`git rev-parse HEAD` -X github.com/FactomProject/factomd/engine.FactomdVersion=`cat ../VERSION`" -v \
    -buildmode=c-shared factomd.go

