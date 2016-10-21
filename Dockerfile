FROM alpine:latest

# Get certificates since we need to talk to AWS on HTTPS.
RUN apk add --update ca-certificates
RUN update-ca-certificates

# Do not run as root.
USER nobody

ADD rdstail .
ENTRYPOINT ["/rdstail"]
