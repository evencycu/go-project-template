mod:
	GO111MODULE=on go build -v
	GO111MODULE=on go mod tidy

test:
	@echo "Start unit tests & vet..."
	go vet $(SOURCE)
	go test -race -cover