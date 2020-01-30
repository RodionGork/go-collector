FROM golang:1.13.7-alpine3.11

EXPOSE 8085

WORKDIR /go/src/app

COPY . .

RUN apk add beanstalkd \
    && apk add protobuf \
    && go get github.com/golang/protobuf/protoc-gen-go

CMD . ./launch-all.sh