#!/usr/bin/env bash

export LOG_LEVEL=warn

i=0
while [[ $i -lt 1000 ]] ; do
    ./go-collector client send
    i=$((i + 1))
done
./go-collector client stats