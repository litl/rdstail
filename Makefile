ifndef GOOS
$(error GOOS is not set)
endif

install:
	go get ./...

lint: install
	go vet github.com/Instamojo/rdstail/...

test: lint
	go test github.com/Instamojo/rdstail/... --cover

build: test
	echo "compiling with `go version`"
	CGO_ENABLED=0 GOOS=$(GOOS) go build -a -installsuffix cgo -o $(GOPATH)/src/github.com/Instamojo/rdstail/rdstail github.com/Instamojo/rdstail
