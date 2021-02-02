export GOPRIVATE=gitlab.com*

build:
	go build

mod: 
	go mod tidy

modvendor:
	- rm go.sum
	GO111MODULE=on go build -v $(SOURCE) 
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor

test:
	@echo "Start unit tests & vet..."
	go vet $(SOURCE)
	go test -race -cover

bench:
	@echo "Start vet & benchmarks..."
	go vet $(SOURCE)
	go test -bench=. -benchtime 5s
