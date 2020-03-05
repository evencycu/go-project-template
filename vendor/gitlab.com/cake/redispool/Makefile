export GOPRIVATE=gitlab.com*

build:
	go build 

test:
	@echo "Start unit tests & vet..."
	go vet $(SOURCE)
	go test -race -cover -timeout 30s $(SOURCE)

clean:
	rm -rf bin pkg
	docker system prune -f
	go mod tidy

modvendor:
	GO111MODULE=on go build -v $(SOURCE) 
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor