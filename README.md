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

0. OS environment setup

    It is suggested to use Linux as your development environment as our production environment is Linux.
    Ubuntu 20.04 LTS is recommended. Other Linux distros like CentOS 8 are also OK.
    The following procedures are based on Ubuntu 20.04 LTS.

    MacOS is also OK. You may need to use Homebrew to install required packages.

    If you are using Windows 10, you are advised to install WSL2.
    Windows OS requirement: Windows 10 version 2004, OS build 19041.264
    https://docs.microsoft.com/en-us/windows/wsl/install-win10
    After setting up WSL2, install Ubuntu 20.04 LTS from Microsoft Store

    Install Go
    Download binary package from https://golang.org/dl/ to /usr/local/
    untar go package and delete the binary package
    Insert the following in your ~/.bashrc
    ```
    export PATH=$PATH:/usr/local/go/bin
    export GO111MODULE=on
    export GOFLAGS='-mod=vendor'
    export GOPRIVATE=gitlab.com
    export NO_PROXY=gitlab.devops.maaii.com
    export no_proxy=gitlab.devops.maaii.com
    export http_proxy=http://192.168.0.30:3128
    ```
    if you are not behind a web proxy server, you don't need last line

    IDE:
    You can use vscode in all your environments.
    If you are using WSL2, you can launch vscode within your Ubuntu Linux. 
    Type "code" in your command prompt.
    Then you can install go extension in vscode marketplace.

    Vim-go is a good choice if you are a hardcore vi user.

    If you have difficulty on setting up your environment, please talk to your team lead.

1. Git setup

    The gitlab account is the LDAP account. You should get your LDAP account from IT team.
    LDAP account is different from Windows AD account, i.e. your email account.
    If you cannot login gitlab, talk to devops team.
    M800 gitlab URL: https://gitlab.devops.maaii.com

    Gitlab configuration
    Generate a SSH key. Guide: https://gitlab.devops.maaii.com/help/ssh/README#generating-a-new-ssh-key-pair
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

    If you have got this go-project-template, you should have probably set up the gitlab account.
    If not, please follow the above procedure to set up the gitlab account.

2. Copy necessary files

    Assumption:
    Assume you are using Linux/MacOS system, bash, find and sed commands should be available.
    Usually they do in a normal Linux/MacOS environment.
    If you are using Windows environment, this procedure doesn't apply. You have to
    read copy.sh to do it manually in Windows environment.

    If your project is named my_go_project, you can alias your project name as mgp.
    The purpose is to fulfill the requirement of this go project template.
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
    $ git remote add origin ssh://git@gitlab.devops.maaii.com:2222/cake/your_project_name.git
    $ git add .
    $ git commit -m "Initial commit"
    $ git push -u origin master
    ```
    remember to change the project path "/cake/your_project_name.git" to your actual project path

    Then you should be able to make the binary
    ```shell
    $ go mod vendor
    $ make
    ```
    
    The binary will be in your ~/go/bin directory

    * replace all error codes in the project alias directory by project error code. 
    (register error code here: [Link](https://issuetracking.maaii.com:9443/pages/viewpage.action?pageId=88354121))  

    Note your default project directory will be in gitlab.com/cake/your_project_name
    Make sure this is the correct project path, if not, please change the path in all related files
    For instance,
    ```shell
    find . -type f -exec sed -i'' -e "s/gitlab.com\/cake\/your_project_name/gitlab.com\/backend\/your_project_name/g" {} +
    ```
    If you are not sure the correct path, please consult your team lead.
    The path is your git path. You should commit your code to gitlab server.
    https://gitlab.com is an alias for ssh://git@gitlab.devops.maaii.com:2222/

    If you want to deploy your project in k8s, which is usually the case,  please review the devops directory. 
    in devops/base/deployment.yaml, make sure the following is correct 
    image: artifactory.maaii.com/lc-docker-local/my_go_project:latest 
    please talk to devops team for the correct path in artifactory server and help you to setup CI/CD pipeline for your application


3. CI/CD guidelines

    Submit your code

    commit your code
    ```shell
    $ git add .
    $ git commit -m "Your commit comment"
    ```

    merge your code with master branch, assume your current branch version is branch_1.0
    ```shell
    $ git checkout master
    $ git pull
    $ git branch -m your_branch_version
    $ git merge branch_1.0
    ```
    If there is a conflict, please resolve the conflict first before you submit it gitlab.
    If there is a conflict, resolve it and commit it again.

    push your code for review.
    ```shell
    $ git push -u origin your_branch_version
    ```
    It will give you a URL for the merge request.
    Go to the link and submit your merge request.
    After that, ask your teammate or team lead to do code review and merge your code to master.

    CI/CD flow
    Your code is merged to gitlab project master branch, jenkins will fetch the new code and build it.
    Jenkins URL: https://emma.devops.m800.com/
    You can login with your LDAP account and search your project.
    Jenkins will do build, unit test, sonarcube code scan and deploy.
    Developers can have access in development and integration environment.
    When you feel comfortable with your latest code, you can tag it and deploy it to integration environment.
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

    Access your build in dev and int environment
    You should get a node port for your application. You need to ask devops team to give you a new node port for your application.
    You cannot randomly assign a node port. They are managed by devops team to avoid conflicts.
    Then you can access your application in URL: http://kube-worker.cloud.m800.com:node_port
    dev and int environments have different node ports.

    After you finish your testing, you can create a RD ticket in Jira to ask application operation team to deploy your application to
    testbed. Then QA can start testing your application. In the RD ticket, make sure you have proper documentation of your build to
    let QA/operation team to understand the changes and bug fix. If the changes are huge, it is suggested to give them a briefing
    to let them understand your application easier.

    Note: You have to develop your unit test/integration code to prove your code is working properly before you submit it to
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
