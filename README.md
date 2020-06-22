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
    Copy completed
    ```

2. Change project name

    replace all `go-project-template` string to `my_project` in all files

3. Change the project info package `gpt`  into your project alias name

    * change folder name

      ```shell
      $ mv gpt mp
      ```

    * replace all `gpt` string to `mp` in all files
    * replace all error codes in `mp` by project error code. (register error code here: [Link](https://issuetracking.maaii.com:9443/pages/viewpage.action?pageId=88354121))  

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
