set -e

export PROJECT_NAME="rdstail"
export IMPORT_PATH="github.com/Instamojo/$PROJECT_NAME"

docker run -v `pwd`:/go/src/$IMPORT_PATH golang:latest bash -c "\
    go version; \
    cd /go/src/$IMPORT_PATH;\
    go get -v -d; \
    export CGO_ENABLED=0; \
    export GOOS=linux; \
    go build -a -installsuffix cgo -o $PROJECT_NAME ."


docker build -t rdstail:latest .
