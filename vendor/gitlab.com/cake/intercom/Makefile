SOURCE=./...
export GOPRIVATE=gitlab.com*
ARTIFACTORY=artifactory.maaii.com/lc-docker-local/
DOCKERTAG=$(ARTIFACTORY)intercom

build: 
	go build -v $(SOURCE) 

test:
	go vet $(SOURCE)
	go test -race -failfast -cover -mod=vendor -timeout 60s .

dockertest:modvendor
	docker build -t $(DOCKERTAG)-test -f Dockerfiletest .
	docker run --network host -v /var/run/docker.sock:/var/run/docker.sock $(DOCKERTAG)-test:latest

clean:
	rm -rf bin pkg
	docker system prune -f
	go mod tidy

modvendor:
	- rm -rf go.sum
	GO111MODULE=on go build -v $(SOURCE) 
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor
