APP=go-project-template
PKGPATH=gitlab.com/cake/gopkg
GOPATH=$(shell env | grep GOPATH | cut -d'=' -f 2)
CONF=local.toml

SKAFFOLD_CONF=devops/skaffold.yaml
SKAFFOLD_DEBUG_CONF=devops/debug/skaffold.yaml
SKAFFOLD_TEMPLATE=devops/skaffold-template.yaml
KUSTOMIZE_CONF=devops/kustomization.yaml
KUSTOMIZE_TEMPLATE=devops/kustomization-template.yaml
BASEDEPLOYMENT=devops/base/deployment.yaml
DEVOPSTOOL=$(GOPATH)/src/gitlab.com/cake/DevOps-Tools

APPCONFIG=$(GOPATH)/src/gitlab.com/cake/app-config
ARTIFACTORY=artifactory.maaii.com/lc-docker-local/
DOCKERTAG=$(ARTIFACTORY)$(APP)
PWD=$(shell pwd)
PORT=$(shell head -10 local.toml | grep port | cut -d'=' -f 2 |tr -d '[:space:]'| tr -d '"')
SOURCE=./...
REVISION=$(shell git rev-list -1 HEAD)
TAG=$(shell git tag -l --points-at HEAD | tail -1)
ifeq ($(TAG),)
TAG=$(REVISION)
endif
BR=$(shell git rev-parse --abbrev-ref HEAD)
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BLUEPRINT_PATH=blueprint/$(APP).apib
export GOPRIVATE=gitlab.com*
SKAFFOLD_SUFFIX=george

build:
	go install -mod=vendor -v -ldflags "-s -X $(PKGPATH).appName=$(APP) -X $(PKGPATH).gitCommit=$(REVISION) -X $(PKGPATH).gitBranch=$(BR) -X $(PKGPATH).appVersion=$(TAG) -X $(PKGPATH).buildDate=$(DATE)" $(SOURCE)

run: build
	@echo "GOPATH: $(GOPATH)"
	$(GOPATH)/bin/$(APP) server --config $(CONF)

test:
	@echo "Start unit tests & vet..."
	go vet $(SOURCE)
	go test -cover -race -timeout 60s $(SOURCE)

clean:
	rm -rf pkg
	go clean --modcache
	docker system prune -f

modrun:
	go install -v $(SOURCE)
	$(GOPATH)/bin/$(APP) server --config $(CONF) 

modvendor:
	- rm go.sum
	go build -mod=mod -v $(SOURCE)
	go mod tidy -compat=1.17
	go mod vendor

m1modvendor:
	- rm go.sum
	go build -mod=mod -tags dynamic -v $(SOURCE)
	go mod tidy
	go mod vendor

mongo:
	docker run -ti --rm -p 27017:27017 --network mongo --name mongo \
		-e MONGO_INITDB_ROOT_USERNAME=admin \
		-e MONGO_INITDB_ROOT_PASSWORD=secret \
		mongo:5.0.6

modcheck:
	go list -mod=mod -m -u all

dockerbuild: modvendor
	docker build --build-arg APPNAME=$(APP) --build-arg GITVERSION=$(TAG) --build-arg GITREVISION=$(REVISION) --build-arg GITBRANCH=$(BR) --platform linux/amd64 -t $(DOCKERTAG) -f devops/Dockerfile .
	
docker: dockerbuild
	docker run -p $(PORT):$(PORT) $(DOCKERTAG):latest 

dockertest: modvendor
	mv .dockerignore .dockerignore1
	docker build -t $(DOCKERTAG)-test -f devops/DockerfileTest .
	docker run --network host -v /var/run/docker.sock:/var/run/docker.sock $(DOCKERTAG)-test:latest
	mv .dockerignore1 .dockerignore

dockerpush: dockerbuild
	docker push $(DOCKERTAG):latest

kustomize: dockerpush
	sed -i '5i \ \ annotations:\n\ \ \ \ revision: $(REVISION)\n\ \ \ \ branch: $(BR)\n\ \ \ \ version: $(TAG)' $(BASEDEPLOYMENT)
	kustomize build $(APPCONFIG)/dev/dev-hk-03/$(APP) | kubectl apply -f -
	sed -i '5,8d' $(BASEDEPLOYMENT)
	kubectl delete po $(shell kubectl get po | grep $(APP) | awk '{print $$1}')

sktemplate:
	@if [ -z "${SKAFFOLD_SUFFIX}" ]; then\
		echo "environment variable: 'SKAFFOLD_SUFFIX' is not set";\
		exit 1;\
	fi
	cp $(SKAFFOLD_TEMPLATE) $(SKAFFOLD_CONF)
	sed -i'' -e 's/SKAFFOLD_SUFFIX/$(SKAFFOLD_SUFFIX)/g' $(SKAFFOLD_CONF)
	- rm $(SKAFFOLD_CONF)-e
	cp $(KUSTOMIZE_TEMPLATE) $(KUSTOMIZE_CONF)
	sed -i'' -e 's/SKAFFOLD_SUFFIX/$(SKAFFOLD_SUFFIX)/g' $(KUSTOMIZE_CONF)
	- rm $(KUSTOMIZE_CONF)-e

skdev: sktemplate modvendor
	skaffold dev -f $(SKAFFOLD_CONF) --trigger manual --port-forward

skrun: sktemplate modvendor
	skaffold run -f $(SKAFFOLD_CONF)

skdelete:
	@echo "Delete skaffold run..."
	skaffold delete -f $(SKAFFOLD_CONF)

skdebug: modvendor
	skaffold dev -f $(SKAFFOLD_DEBUG_CONF) --trigger manual

apib:
	@echo "Make sure you have install snowboard"
	snowboard lint blueprint.md
	snowboard apib -o blueprint/$(APP).apib blueprint.md
	sed -i'' -e 's/LASTUPDATED/$(DATE)/g' $(BLUEPRINT_PATH)
	- rm $(BLUEPRINT_PATH)-e
	@echo "Completed"

apibrun:
	@echo "Make sure you have install snowboard"
	snowboard --watch --watch-interval 2s html -o blueprint.html -s blueprint.md

sonarscan:
	$(DEVOPSTOOL)/sonar-scanner-tools/local-scan.sh test $(PWD)
	$(DEVOPSTOOL)/sonar-scanner-tools/local-scan.sh upload $(PWD)

adrtoc:
	adr generate toc > doc/adr/README.md

adrview:
	adr-viewer --serve
