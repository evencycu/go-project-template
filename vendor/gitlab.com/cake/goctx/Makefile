build:
	go build

test:
	go test -race -cover -timeout=15s

modvendor:
	- rm go.sum
	GO111MODULE=on go build -v $(SOURCE) 
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor