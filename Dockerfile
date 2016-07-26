FROM golang:1.7

RUN mkdir -p /go/src/rdstail
WORKDIR /go/src/rdstail

COPY . /go/src/rdstail
RUN go-wrapper download
RUN go-wrapper install
