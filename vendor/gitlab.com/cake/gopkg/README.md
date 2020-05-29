# Introduction

The common errors, shared utilities of Golang based microservices.

## Requirement

* `Golang v1.8.0+`

## TODO

* utils/slack

## Project Structure

```sh
├── CHANGELOG.md
├── README.md
├── errors_test.go
├── errors.go
├── init.go
└── version.go
```

## The concept about error

### CodeError

The basic error type of production usage (interface). We need one more `Error Code` in most use case. Therefore, we define the `CodeError` interface implements not only the `Error()` but also `ErrorCode()`.

* Methods
  * Error() - error message as standard `error`
  * ErrorCode() - error code in **7** digits

### AsCodeError

* According to the name, you will know it could be handle as Code Error
* But it was created to support more actually.
* The Implemented struct type is TraceCodeError.
* Interface Detail
* AsEqual - It should contain a CodeError, AsEqual should be used in error handling.

## Version information

### Project version injection at build time

With golang `-ldflags -X`, we can inject version information at build time.

```Makefile
  APP=YOUR_APP_NAME
  PKGPATH=gitlab.com/cake/gopkg
  REVISION=$(shell git rev-list -1 HEAD)
  TAG=$(shell git tag -l --points-at HEAD | tail -1)
  ifeq ($(TAG),)
  TAG=$(REVISION)
  endif
  BR=$(shell git rev-parse --abbrev-ref HEAD)
  DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

  build:
    go install -i -v -ldflags "-s -X $(PKGPATH).appName=$(APP) -X $(PKGPATH).gitCommit=$(REVISION) -X $(PKGPATH).gitBranch=$(BR) -X $(PKGPATH).appVersion=$(TAG) -X $(PKGPATH).buildDate=$(DATE)"
```

You can add a handler to show the version information.

```go
import (
  "net/http"

  "github.com/gin-gonic/gin"
  "gitlab.com/cake/gopkg"
)

func version(c *gin.Context) {
  c.JSON(http.StatusOK, gopkg.GetVersion())
}
```

And it will show result like following:

```json
{
  "version": "v0.1.2",
  "gitCommit": "4384d0ea4a6cf0615dedd13042412739e8812ab8",
  "gitBranch": "gopkg",
  "buildDate": "2020-06-01T03:42:12Z",
  "goOs": "darwin",
  "goArch": "amd64",
  "goVer": "go1.14.2"
}
```


### Version metrics

```prometheus
# HELP go_build_info Build information about the main Go module.
# TYPE go_build_info gauge
go_build_info{path="gitlab.com/cake/gopkg",version="v0.1.2"} 1


# HELP go_mod_info A metric with a constant '1' value labeled by dependency name, version, from which go-project-template was built.
# TYPE go_mod_info gauge
go_mod_info{name="github.com/FZambia/sentinel",program="go-project-template",version="v1.1.0"} 1
go_mod_info{name="github.com/beorn7/perks",program="go-project-template",version="v1.0.1"} 1
go_mod_info{name="github.com/cespare/xxhash/v2",program="go-project-template",version="v2.1.1"} 1
go_mod_info{name="github.com/confluentinc/confluent-kafka-go",program="go-project-template",version="v1.1.0"} 1
go_mod_info{name="github.com/eaglerayp/go-conntrack",program="go-project-template",version="v0.1.1"} 1
go_mod_info{name="github.com/fsnotify/fsnotify",program="go-project-template",version="v1.4.7"} 1
go_mod_info{name="github.com/gin-contrib/sse",program="go-project-template",version="v0.1.0"} 1
go_mod_info{name="github.com/gin-gonic/gin",program="go-project-template",version="v1.6.3"} 1
go_mod_info{name="github.com/globalsign/mgo",program="go-project-template",version="v0.0.0-20181015135952-eeefdecb41b8"} 1
go_mod_info{name="github.com/go-playground/locales",program="go-project-template",version="v0.13.0"} 1
go_mod_info{name="github.com/go-playground/universal-translator",program="go-project-template",version="v0.17.0"} 1
go_mod_info{name="github.com/go-playground/validator/v10",program="go-project-template",version="v10.2.0"} 1
go_mod_info{name="github.com/gofrs/uuid",program="go-project-template",version="v3.2.0+incompatible"} 1
go_mod_info{name="github.com/golang/protobuf",program="go-project-template",version="v1.4.0"} 1
go_mod_info{name="github.com/gomodule/redigo",program="go-project-template",version="v2.0.0+incompatible"} 1
go_mod_info{name="github.com/hashicorp/hcl",program="go-project-template",version="v1.0.0"} 1
go_mod_info{name="github.com/jpillora/backoff",program="go-project-template",version="v1.0.0"} 1
go_mod_info{name="github.com/leodido/go-urn",program="go-project-template",version="v1.2.0"} 1
go_mod_info{name="github.com/magiconair/properties",program="go-project-template",version="v1.8.1"} 1
go_mod_info{name="github.com/mattn/go-isatty",program="go-project-template",version="v0.0.12"} 1
go_mod_info{name="github.com/matttproud/golang_protobuf_extensions",program="go-project-template",version="v1.0.1"} 1
go_mod_info{name="github.com/mitchellh/mapstructure",program="go-project-template",version="v1.1.2"} 1
go_mod_info{name="github.com/opentracing/opentracing-go",program="go-project-template",version="v1.1.0"} 1
go_mod_info{name="github.com/pelletier/go-toml",program="go-project-template",version="v1.3.0"} 1
go_mod_info{name="github.com/pkg/errors",program="go-project-template",version="v0.8.1"} 1
go_mod_info{name="github.com/povilasv/prommod",program="go-project-template",version="v0.0.12"} 1
go_mod_info{name="github.com/prometheus/client_golang",program="go-project-template",version="v1.6.0"} 1
go_mod_info{name="github.com/prometheus/client_model",program="go-project-template",version="v0.2.0"} 1
go_mod_info{name="github.com/prometheus/common",program="go-project-template",version="v0.9.1"} 1
go_mod_info{name="github.com/prometheus/procfs",program="go-project-template",version="v0.0.11"} 1
go_mod_info{name="github.com/sirupsen/logrus",program="go-project-template",version="v1.4.2"} 1
go_mod_info{name="github.com/spf13/afero",program="go-project-template",version="v1.2.2"} 1
go_mod_info{name="github.com/spf13/cast",program="go-project-template",version="v1.3.0"} 1
go_mod_info{name="github.com/spf13/cobra",program="go-project-template",version="v0.0.6"} 1
go_mod_info{name="github.com/spf13/jwalterweatherman",program="go-project-template",version="v1.1.0"} 1
go_mod_info{name="github.com/spf13/pflag",program="go-project-template",version="v1.0.3"} 1
go_mod_info{name="github.com/spf13/viper",program="go-project-template",version="v1.5.0"} 1
go_mod_info{name="github.com/subosito/gotenv",program="go-project-template",version="v1.2.0"} 1
go_mod_info{name="github.com/uber/jaeger-client-go",program="go-project-template",version="v2.20.1+incompatible"} 1
```
