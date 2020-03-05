# Go Project Template

## M800 Libraries

* `gitlab.com/cake/goctx`
* `gitlab.com/cake/gotrace/v2`
* `gitlab.com/cake/m800log`
* `gitlab.com/cake/mgopool`

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

## How to create a new project

1. Copy necessary files

    ```shell
    $ ./copy.sh ../my_project
    copy completed
    ```

2. Change project name

    replace all `go-project-template` string to `my_project` in all files

3. Change the project info package `gpt`  into your project alias name

    ```shell
    $ mv gpt mp
    ```

    replace all `gpt` string to `mp` in all files
