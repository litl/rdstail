FROM golang:latest as builder
WORKDIR /go/src/github.com/Instamojo/rdstail/
COPY . .
RUN go get -v -d
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o rdstail .

####################
FROM alpine:latest

# Get CA certificates since we need to talk to AWS on HTTPS.
RUN apk --no-cache add ca-certificates
RUN update-ca-certificates

# Copy the binary from build image
COPY --from=builder /go/src/github.com/Instamojo/rdstail .

# Do not run as root.
USER nobody

ENTRYPOINT ["/rdstail"]
