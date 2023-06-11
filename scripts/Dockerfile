FROM golang:1.19-alpine

WORKDIR /opt/gpw

RUN echo "ulimit is $(ulimit -n)" && \
    ulimit -n 1024000 && \
    apk add git && \
    apk add openssh && \
    apk add gcc && \
    apk add musl-dev && \
    apk add build-base && \
    apk add bash

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn

COPY . .
