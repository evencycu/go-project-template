# docker build -t [image_name]:[tag] .
FROM golang:1.11.5-alpine

RUN apk --no-cache add --update ca-certificates bash 
WORKDIR /go/src/gitlab.com/liveconnect/lc-chatbot-agent
EXPOSE 8999
ENV GO111MODULE=on
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go install -v -mod=vendor ./...
CMD lc-chatbot-agent -config ./local.toml