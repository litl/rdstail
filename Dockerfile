FROM alpine:3.3

RUN apk upgrade --update && \
  apk add --update curl ca-certificates

COPY rdstail run.sh /app/

WORKDIR /app

ENTRYPOINT [ "/app/run.sh" ]
