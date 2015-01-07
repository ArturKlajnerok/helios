#!/bin/bash

if [[ $# -eq 0 ]] 
then
	echo "Provide one of possible options: all, unit, e2e"
	exit 1
fi

run_tests() {
    godep go test $1 github.com/Wikia/helios/models
    godep go test $1 github.com/Wikia/helios/helios
}


if [[ $1 == "all" ]]; then
    run_tests
fi

if [[ $1 == "unit" ]]; then
    run_tests -short
fi
