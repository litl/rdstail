FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/pperzyna/rdstail/
COPY . .
RUN GO111MODULE=on go mod download && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/rdstail

FROM scratch
COPY --from=builder /go/bin/rdstail /go/bin/rdstail
ENTRYPOINT ["/go/bin/rdstail"]