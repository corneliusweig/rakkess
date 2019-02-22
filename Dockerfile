FROM golang:alpine

RUN apk add make git

RUN mkdir -p /go/src/github.com/corneliusweig/rakkess/

WORKDIR /go/src/github.com/corneliusweig/rakkess/

CMD git clone --depth 1 https://github.com/corneliusweig/rakkess.git . && \
    make all && \
    mv out/* /go/bin
