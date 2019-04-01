# docker build -t [image_name]:[tag] .
FROM golang:1.11.6-alpine as builder
WORKDIR /go/src/
ENV GO111MODULE=on

# Copy vendor first and build for cache layers
COPY go.mod .
COPY go.sum .
COPY vendor vendor
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install -mod=vendor -ldflags="-w -s" -v ./vendor/...

# Copy source codes and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install -mod=vendor -ldflags="-w -s" -v ./...

# For minimal image
FROM alpine:3.9
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/go-project-template .
ENTRYPOINT ["./go-project-template"]
