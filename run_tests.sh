#!/bin/bash

if [[ $# -eq 0 ]] 
then
	echo "Provide one of possible options: all, unit, e2e"
	exit 1
fi

run_unit_tests() {
    godep go test github.com/Wikia/helios/models
}

run_e2e_tests() {
    godep go test github.com/Wikia/helios/e2e
}


if [[ $1 == "all" || $1 == "unit" ]]; then
    run_unit_tests
fi

if [[ $1 == "all" || $1 == "e2e" ]]; then
    run_e2e_tests
fi
