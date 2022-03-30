FROM artifactory.maaii.com/lc-docker-local/golang:1.17.8-alpine-kafka as builder
WORKDIR /go/src/

# Copy source codes and build
COPY . .
ARG APPNAME
ARG GITVERSION
ARG GITBRANCH
ARG GITREVISION
ENV PKGPATH=gitlab.com/cake/gopkg
RUN export BUILDDATE=$(date +"%Y-%m-%dT%H:%M:%SZ") && GOOS=linux GOARCH=amd64 go install -tags dynamic -tags musl -mod=vendor -gcflags "all=-N -l" -ldflags="-X $PKGPATH.appName=$APPNAME -X $PKGPATH.gitCommit=$GITREVISION -X $PKGPATH.appVersion=$GITVERSION -X $PKGPATH.buildDate=$BUILDDATE -X $PKGPATH.gitBranch=$GITBRANCH" -v ./...

# For minimal image
FROM artifactory.maaii.com/lc-docker-local/golang:1.17.8-alpine-delve
COPY --from=builder /go/bin/go-project-template .

# Init server
COPY local.toml .
COPY go.mod .
CMD /go/bin/dlv --listen=:2345 --headless=true --log=true --log-output=debugger --accept-multiclient --api-version=2 exec /go-project-template -- server --config ./local.toml
