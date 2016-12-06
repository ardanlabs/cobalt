.PHONY: build test 
test:
	go test -v github.com/ardanlabs/cobalt

build:
	go clean -i github.com/ardanlabs/cobalt 
	go build github.com/ardanlabs/cobalt 
	go vet github.com/ardanlabs/cobalt 
