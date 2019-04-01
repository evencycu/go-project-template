# Template

## M800 Libraries

* `gitlab.com/general-backend/goctx`
* `gitlab.com/general-backend/gotrace`
* `gitlab.com/general-backend/m800log`
* `gitlab.com/general-backend/mgopool`

## Common Libraries

* `github.com/spf13/viper`
* `github.com/uber/jaeger-client-go`
* `github.com/uber/jaeger-client-go/config`
* `github.com/zsais/go-gin-prometheus`

## Base functions

* log (m800log, high-level wrapper of logrus)
* config (viper)
* tracing (gotrace, high-level wrapper of jaeger)
* metric (gin-prometheus)
* mongodb (mgopool, high-level wrapper of mgo)

## CICD integration

* make sure all the variable config in different region patch in `app-configuration`
