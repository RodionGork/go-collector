#!/bin/sh

beanstalkd -p 11300 > /dev/null &

protoc --go_out=. *.proto

go build

./go-collector server &

httpClientPort=8085 ./go-collector http-client > /dev/null