FROM golang:1.7

WORKDIR /go/src/github.com/Instamojo/rdstail

COPY . /go/src/github.com/Instamojo/rdstail
RUN go-wrapper download
RUN go-wrapper install
