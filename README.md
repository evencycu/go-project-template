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

    if your project is named my_go_project, you can alias your project name as mgp 

    ```shell
    $ ./copy.sh my_go_project mgp
    Copy completed
    ```
    your project directory will be created on upper directory relative to this go-project-template

    cd to your project directory
    ```shell
    $ git init
    $ go mod vendor
    ```
    if you want to use kafka in your go project, you have to install librdkafka-dev if you are using Debian/Ubuntu
    ```shell
    $ sudo apt install librdkafka-dev
    ```
    if you don't use kafka in your project, remove kafka related content in go.mod and go.sum

    Then you should be able to make the binary
    ```shell
    $ make
    ```
    
    The binary will be in your ~/go/bin directory

    * replace all error codes in the project alias directory by project error code. (register error code here: [Link](https://issuetracking.maaii.com:9443/pages/viewpage.action?pageId=88354121))  

    Note your default project directory will be in gitlab.com/cake/your_project_name
    Make sure this is the correct project path, if not, please change the path in all related files
    For instance,
    ```shell
    find . -type f -exec sed -i'' -e "s/gitlab.com\/cake\/your_project_name/gitlab.com\/backend\/your_project_name/g" {} +
    ```
    If you are not sure the correct path, please consult your team lead.

    if you want to deploy your project in k8s, please review the devops directory. 
    in devops/base/deployment.yaml, make sure the following is correct 
    image: artifactory.maaii.com/lc-docker-local/george-project:latest 
    please talk to devops team for the correct path and help your to setup CI/CD pipeline for your application

## How to write Architecture Decision Records

### Requirements

[Install adr-tools](https://github.com/npryce/adr-tools/blob/master/INSTALL.md)

[Install adr viewer (optional)](https://github.com/mrwilson/adr-viewer)

### Examples

Please check [doc/adr](doc/adr/) for all examples.

```tree
doc/adr
├── 0001-record-architecture-decisions.md
├── 0002-adr-format-with-madr.md
├── 0003-adr-format-with-lightweight-adr.md
├── 0004-adr-support-tool.md
└── README.md
```

### Usage

Create an ADR directory in the root of your project:

```shell
  adr init doc/adr
```

Create Architecture Decision Records

```shell
  adr new Implement as Unix shell scripts
```

To create a new ADR that supersedes a previous one (ADR 9, for example), use the -s option.

```shell
  adr new -s 9 Use Rust for performance-critical functionality
```

To create a new ADR that amend a previous one (ADR 10, for example), use the -l option.

```shell
  adr new -l "10:Amends:Amended by"  Use SSD to improve performance
```

For further information, use the built in help:

```shell
  adr help
```

For create web UI, use `adr-viewer` and access it via `localhost:8000` :

```shell
  adr-viewer --serve
```
