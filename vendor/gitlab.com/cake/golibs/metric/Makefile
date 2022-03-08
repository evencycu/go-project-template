SOURCE=./...
export GOPRIVATE=gitlab.com*

test:
	@echo "Start unit tests & vet..."
	go vet $(SOURCE)
	go test -race -cover $(SOURCE)

run-origin:
	go run example/gin-metrics-origin/main.go

run-custom:
	go run example/gin-metrics-custom/main.go

run-trace:
	go run example/gin-metrics-trace/main.go

modvendor:
	GO111MODULE=on go build -mod=mod -v $(SOURCE) 
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor

