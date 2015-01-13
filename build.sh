#!/bin/bash

OUTPUT=$GOPATH/bin/helios/helios

if [ $1 ]; then
OUTPUT=$1
fi;

$GOPATH/bin/godep go build -o $OUTPUT github.com/Wikia/helios
