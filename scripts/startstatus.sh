#!/usr/bin/env bash
if [[ -z $1 ]]; then
file=out.txt
else
file=$1
fi

reset
tail -n +0 -f $file | gawk -f $GOPATH/src/github.com/FactomProject/factomd/scripts/status.awk
