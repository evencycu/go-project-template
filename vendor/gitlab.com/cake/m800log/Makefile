build:
	go build

mod: 
	go mod tidy

modvendor:
	GO111MODULE=on go build -v $(SOURCE) 
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor

test:
	@echo "Start unit tests & vet..."
	go vet $(SOURCE)
	go test -race -cover