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

    Assumption:
    Assume you are using Linux/MacOS system, bash, find and sed commands are avaiable.
    Usualy they do in a normal Linux/MacOS environment.
    If you are using Windows environment, this procedure doesn't apply. You have to
    read copy.sh to do it manually in Windows environment.

    If your project is named my_go_project, you can alias your project name as mgp.
    The purpose is to fulfill the requriement of this go project template.
    The project alias will be a go module.

    ```shell
    $ ./copy.sh my_go_project mgp
    ```
    your project directory will be created on upper directory relative to this go-project-template

    cd to your project directory

    If you want to use kafka in your go project, you have to install librdkafka-dev if you are using Debian/Ubuntu
    If you are using CentOS/RHEL, you have to install librdkafka-devel
    If you don't use kafka in your project, remove kafka related content in go.mod and go.sum and newKafkaProducerConfig function in command/server.go
    and skip installing librdkafka
    Debian/Ubuntu
    ```shell
    $ sudo apt install librdkafka-dev
    ```
    CentOS/RHEL
    ```shell
    $ sudo yum install librdkafka-devel
    ```

    ```shell
    $ git init
    $ go mod vendor
    ```

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
    The path is your git path. You should commit your code to gitlab server.
    https://gitlab.com is an alias for ssh://git@gitlab.devops.maaii.com:2222/

    If you want to deploy your project in k8s, which is usualy the case,  please review the devops directory. 
    in devops/base/deployment.yaml, make sure the following is correct 
    image: artifactory.maaii.com/lc-docker-local/my_go_project:latest 
    please talk to devops team for the correct path in artifactory server and help you to setup CI/CD pipeline for your application


2. CI/CD guidelines

    The gitlab acocunt is the LDAP account. You should get the LDAP account from IT team.
    If you cannot login gitlab, talk to devops team.
    gitlab URL: https://gitlab.devops.maaii.com

    Gitlab configuration
    Then generate a SSH key. Guide: https://gitlab.devops.maaii.com/help/ssh/README#generating-a-new-ssh-key-pair
    Deploy the SSH key in your gitlab account.
   
    Git configuration
    Please change user.name and user.email. If you are using proxy server in office, please setup http.proxy.
    ```shell
    $ git config --global url."ssh://git@gitlab.devops.maaii.com:2222/".insteadOf https://gitlab.com/
    $ git config --global user.name "John Doe"
    $ git config --global user.email johndoe@m800.com
    $ git config --global core.editor vi
    $ git config --global http.proxy http://192.168.0.30:3128
    ```
    Git tutorial
    https://git-scm.com/docs/gittutorial
    https://github.com/twtrubiks/Git-Tutorials

    If you have get this go-project-template, you should have probably set up the gitlab account.
    If not, please follow the above procedure to set up the gitlab acocunt.

    CI/CD flow
    when you push your code to gitlab project master branch, jenkins will fetch the new code and build it.
    Jenkins URL: https://emma.devops.m800.com/
    You can login with your LDAP account and search your project.
    Jenkins will do build, unit test, sonarcube code scan and deploy.
    Developers can have access in development and integration environment.
    When you feel comforable with your latest code, you can tag it and deploy it to integration environment.
    ```shell
    $ git tag -v v.0.0.1
    $ git push origin v.0.0.1
    ```
    Then you can login jenkins. In your Jenkins project, go to "Build with Parameters", then you can deploy your code to
    dev and int environment.

    In order to make sure jenkins can build your code and deploy your build successfully, you have adjust your devops
    directory.
    Check files in devops/base/. Details of adjustment is out of scope of this documentation. Please consult devops team
    if you have any issue. If your files in devops/base/ are not properly set up, jenkins cannot build your code successfully.

    Access your buid in dev and int environment
    You should get a node port for your application. You need to ask devops team to give you a new node port for your application.
    Then you can access your application in URL: http://kube-worker.cloud.m800.com:node_port
    dev and int environments have different node ports.

    After you finish your testing, you can create a RD ticket in Jira to ask application operation team to deploy your application to
    testbed. Then QA can start testing your application. In the RD ticket, make sure you have proper documentation of your build to
    let QA/operation team to understand the changes and bug fix. If the changes are huge, it is suggested to give them a briefing
    to let them understand your application easier.

    Note: You have to develop your unit test/integration code to prove your code is working properly before you can submit it to
    QA for testing.

3. Production issue

    Application Operation team(AO) is responsible for production support. If there is a production issue, they are the team to do the
    troubleshooting. Make sure your code have enough logs to let AO team to understand what happens in your application.
    If they cannot do the troubleshooting task, you are responsible for it. It will kill you a lot of time.
    For logging guideline, please contact your team lead. This is very important. 


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
