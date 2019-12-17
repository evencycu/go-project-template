build:
	go build

test: 
	go test -cover -race -v ./...

mod: 
	go mod tidy

modvendor:
	GO111MODULE=on go build -v $(SOURCE) 
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor

dockertest:modvendor
	docker build -t gopkg-test -f dockerfiletest .
	docker run --network host -v /var/run/docker.sock:/var/run/docker.sock gopkg-test:latest
