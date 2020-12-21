export GOPRIVATE=gitlab.com*

build:
	go build

test: 
	go test -cover -race -v ./...

mod: 
	go mod tidy

modvendor:
	- rm go.sum
	GO111MODULE=on go build -v $(SOURCE) 
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor