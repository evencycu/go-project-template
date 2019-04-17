APP=go-project-template
PKGPATH=gitlab.com/rayshih/go-project-template/gpt
CONF=local.toml
SKAFFOLD_CONF=devops/skaffold.yaml
BASEDEPLOYMENT=devops/base/deployment.yaml
ARTIFACTORY=artifactory.devops.maaii.com/lc-docker-local/
DOCKERTAG=$(ARTIFACTORY)$(APP)
PWD=$(shell pwd)
PORT=$(shell head -10 local.toml | grep port | cut -d'=' -f 2 |tr -d '[:space:]'| tr -d '"')
SOURCE=./...
GOPATH=$(shell env | grep GOPATH | cut -d'=' -f 2)
REVISION=$(shell git rev-list -1 HEAD)
TAG=$(shell git tag -l --points-at HEAD)
ifeq ($(TAG),)
TAG=$(REVISION)
endif
BR=$(shell git rev-parse --abbrev-ref HEAD)
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

run: build
	$(GOPATH)/bin/$(APP) -config $(CONF)

update:
	git pull

build: 
	go install -i -v -ldflags "-s -X $(PKGPATH).gitCommit=$(REVISION) -X $(PKGPATH).appVersion=$(TAG) -X $(PKGPATH).buildDate=$(DATE)" $(SOURCE) 

test:
	@echo "Start unit tests & vet..."
	go vet $(SOURCE)
	go test -race -cover $(SOURCE)

clean:
	rm -rf bin pkg
	docker rmi $(shell docker images | grep none | awk '{print $$3}')

modrun:
	GO111MODULE=on go install -mod=vendor -v $(SOURCE)
	$(GOPATH)/bin/$(APP) -config $(CONF) 

modvendor:
	GO111MODULE=on go build -v $(SOURCE) 
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor

dockerbuild:modvendor
	docker build --build-arg GITVERSION=$(TAG) --build-arg GITREVISION=$(REVISION) --build-arg GITBRANCH=$(BR) -t $(DOCKERTAG) -f devops/Dockerfile .

dockerrun:dockerbuild
	docker run -p $(PORT):$(PORT) $(DOCKERTAG):latest

dockerpush:dockerbuild
	docker push $(DOCKERTAG):latest

kustomize:dockerpush
	sed -i '5i \ \ annotations:\n\ \ \ \ revision: $(REVISION)\n\ \ \ \ branch: $(BR)\n\ \ \ \ version: $(TAG)' $(BASEDEPLOYMENT)
	kustomize build devops/dev/ | kubectl apply -f -
	sed -i '5,8d' $(BASEDEPLOYMENT)
	kubectl delete po $(shell kubectl get po | grep $(APP) | awk '{print $$1}')

skdev:	modvendor
	skaffold dev -f $(SKAFFOLD_CONF) --trigger manual

skrun:	modvendor
	skaffold run -f $(SKAFFOLD_CONF)

skdelete:
	@echo "Delete skaffold run..."
	skaffold delete -f $(SKAFFOLD_CONF)