APP=lc-chatbot-agent
CONF=local.toml
SKAFFOLD_CONF=devops/skaffold.yaml
PWD=$(shell pwd)
PORT=$(shell head -3 conf.d/local.toml | grep port | cut -d'=' -f 2 |tr -d '[:space:]'| tr -d '"')
SOURCE=./...
GOPATH=$(shell env | grep GOPATH | cut -d'=' -f 2)
export GO111MODULE=on

initrun:
	go run ./

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

skbuild:
	@echo "Start skaffold build..."
	skaffold build -f $(SKAFFOLD_CONF)

skrun:
	@echo "Start skaffold build..."
	skaffold run -f $(SKAFFOLD_CONF)

skdev:
	@echo "Start skaffold build..."
	skaffold dev -f $(SKAFFOLD_CONF) --trigger manual

skdelete:
	@echo "Start skaffold build..."
	skaffold delete -f $(SKAFFOLD_CONF)