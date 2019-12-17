SOURCE=./...

build: 
	go build -v $(SOURCE) 

test:
	go vet $(SOURCE)
	go test -race -cover -mod=vendor -timeout 60s .

clean:
	rm -rf bin pkg
	docker system prune -f
	go mod tidy

modvendor:
	rm -rf go.sum
	GO111MODULE=on go build -v $(SOURCE) 
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor
