APP=go-project-template
PKGPATH=gitlab.com/rayshih/go-project-template/gpt
CONF=local.toml
SKAFFOLD_CONF=devops/skaffold.yaml
BASEDEPLOYMENT=devops/base/deployment.yaml
APPCONFIG=$(GOPATH)/src/gitlab.com/cake/app-config
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
	rm -rf pkg
	go clean --modcache
	docker system prune -f

modrun:
	GO111MODULE=on go install -mod=vendor -v $(SOURCE)
	$(GOPATH)/bin/$(APP) -config $(CONF) 

modvendor:
	GO111MODULE=on go build -v $(SOURCE) 
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor
	
dockerbuild:modvendor
	docker build --build-arg GITVERSION=$(TAG) --build-arg GITREVISION=$(REVISION) --build-arg GITBRANCH=$(BR) -t $(DOCKERTAG) -f devops/Dockerfile .
	
docker:dockerbuild
	docker run -p $(PORT):$(PORT) $(DOCKERTAG):latest 

dockertest:modvendor
	mv .dockerignore .dockerignore1
	docker build -t $(DOCKERTAG)-test -f devops/DockerfileTest .
	docker run --network host -v /var/run/docker.sock:/var/run/docker.sock $(DOCKERTAG)-test:latest
	mv .dockerignore1 .dockerignore

dockerpush:dockerbuild
	docker push $(DOCKERTAG):latest

kustomize:dockerpush
	sed -i '5i \ \ annotations:\n\ \ \ \ revision: $(REVISION)\n\ \ \ \ branch: $(BR)\n\ \ \ \ version: $(TAG)' $(BASEDEPLOYMENT)
	kustomize build $(APPCONFIG)/dev/dev-hk-03/$(APP) | kubectl apply -f -
	sed -i '5,8d' $(BASEDEPLOYMENT)
	kubectl delete po $(shell kubectl get po | grep $(APP) | awk '{print $$1}')

skdev:	modvendor
	skaffold dev -f $(SKAFFOLD_CONF) --trigger manual

skrun:	modvendor
	skaffold run -f $(SKAFFOLD_CONF)

skdelete:
	@echo "Delete skaffold run..."
	skaffold delete -f $(SKAFFOLD_CONF)

apib:
	@echo "Make sure you have install snowboard"
	snowboard lint blueprint.md
	# snowboard html -o blueprint/blueprint.html blueprint.md
	snowboard apib -o blueprint/blueprint.apib blueprint.md
	sed -i 's/XVERSION/$(TAG)/g' blueprint/blueprint.apib
	sed -i 's/XGITCOMMIT/$(REVISION)/g' blueprint/blueprint.apib
	sed -i 's/XLASTUPDATED/$(DATE)/g' blueprint/blueprint.apib

apibrun: 
	@echo "Make sure you have install snowboard"
	snowboard --watch --watch-interval 2s html -o blueprint.html -s blueprint.md
