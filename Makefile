APP=lc-chatbot-agent
CONF=conf.d/local.toml
PWD=$(shell pwd)
VER=$(shell head -3 CHANGELOG.md |grep v |awk '{ print $2; }')
PORT=$(shell head -3 conf.d/local.toml | grep port | cut -d'=' -f 2 |tr -d '[:space:]'| tr -d '"')
SOURCE=./...
GOPATH=$(shell env | grep GOPATH | cut -d'=' -f 2)
export GO111MODULE=on

update:
	git pull

build: 
	go install -mod=vendor -v $(SOURCE) 

test:
	@echo "Start unit tests & vet..."
	go vet $(SOURCE)
	go test -race -cover $(SOURCE)

run: build
	$(GOPATH)/bin/$(APP) -config $(CONF)

clean:
	rm -rf bin pkg

vendor:
	go build -v $(SOURCE) 
	go mod tidy
	go mod vendor

docker:
	docker build -t $(APP) .
	docker run -p $(PORT):$(PORT) $(APP):latest

kustomize:
	kustomize build devops/app-configuration/$(APP)/dev/dev-hk-01/