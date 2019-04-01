# CICD integration

## Prerequisite

* Docker version `18.+`
* Skaffold `v0.26.0`

  ```bash
  # Linux
  curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/v0.26.0/skaffold-linux-amd64 && chmod +x skaffold && sudo mv skaffold /usr/local/bin

  # macOS
  curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/v0.26.0/skaffold-darwin-amd64 && chmod +x skaffold && sudo mv skaffold /usr/local/bin

  # Windows
  https://storage.googleapis.com/skaffold/releases/v0.26.0/skaffold-windows-amd64.exe
  ```

* Kustomize `v1.0.11`

  ```bash
  # Linux
  https://github.com/kubernetes-sigs/kustomize/releases/download/v1.0.11/kustomize_1.0.11_linux_amd64

  # macOS
  https://github.com/kubernetes-sigs/kustomize/releases/download/v1.0.11/kustomize_1.0.11_darwin_amd64

  # Windows
  https://github.com/kubernetes-sigs/kustomize/releases/download/v1.0.11/kustomize_1.0.11_windows_amd64.exe
  ```

## Build docker images

  There are some tricks in `Dockerfile` and `.dockerignore` that helps build docker images faster.

* Use two-stage build in `Dockerfile`, and copy the less change files first. For example, copy the vendor directory first in Golang project. It will help to create some cached layers. (To use two-stage Dockerfile, Docker version 18.00+ is required)

* Add `.dockerignore` file. Write down anything that is not related to generate a build. For example: `devops` directory, `.git` directory, markdown files, binaries and so on.

## Skaffold

[Github repository](https://github.com/GooglecontainerTools/skaffold)

Skaffold is a command line tool that facilitates continuous development for Kubernetes applications. Skaffold is the easiest way to share your project with the world: `git clone` and `skaffold run`

### How to use it

* `skaffold run -f devops/skaffold.yaml`: Build image, push image and deploy to K8s.

* `skaffold dev -f devops/skaffold.yaml --trigger manual`: Develop mode. Do the all steps as `skaffold run`. Also port-forward pods to local with random port, press any key to rebuild/redeploy the changes.

* `skaffold delete -f devops/skaffold.yaml`: Delete all resources which `skaffold run` will deploy.

### Note

* **Always** use `latest` tag to save docker registry spaces. If you need to know what commit is in current deployment, add commit information in Kubernetes `annotation`.

## Troubleshooting

1. `build artifact: Error parsing reference: "golang:1.11.6-stretch as builder" is not a valid repository/tag:invalid reference format`

    ![sample](../img/docker_version_issue01.jpg)

    A: Update Docker version to `18.+` and restart docker daemon