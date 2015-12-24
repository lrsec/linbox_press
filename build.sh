#!/usr/bin/env bash
cd ../..
export GOPATH=`pwd`:$GOPATH

# dependency
go get github.com/cihub/seelog

# clean
cd bin
rm -rf ./*

# build
cd ../src/linbox_stress
go install

# copy
cp ./config/* ../../bin/