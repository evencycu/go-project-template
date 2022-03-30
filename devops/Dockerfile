FROM artifactory.maaii.com/lc-docker-local/golang:1.17.8-alpine-kafka as builder
WORKDIR /go/src/

# Copy source codes and build
COPY . .
ARG APPNAME
ARG GITVERSION
ARG GITBRANCH
ARG GITREVISION
ENV PKGPATH=gitlab.com/cake/gopkg
RUN export BUILDDATE=$(date +"%Y-%m-%dT%H:%M:%SZ") && CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go install -tags dynamic -tags musl -mod=vendor -ldflags="-w -s -X $PKGPATH.appName=$APPNAME -X $PKGPATH.gitCommit=$GITREVISION -X $PKGPATH.appVersion=$GITVERSION -X $PKGPATH.buildDate=$BUILDDATE -X $PKGPATH.gitBranch=$GITBRANCH" -v ./...

# For minimal image
FROM alpine:edge
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/go-project-template .

# Init server
COPY local.toml .
COPY go.mod .
CMD ./go-project-template server --config ./local.toml
