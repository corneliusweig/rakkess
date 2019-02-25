# rakkess
Review Access - kubectl plugin to show an access matrix for all available resources

## Intro
Have you ever wondered what access rights you have on a provided kubernetes cluster?
For single resources you can use `kubectl auth can-i list deployments`, but maybe you are looking for a complete overview?
This is what `rakkess` is for.
It lists access rights for the current user for all server resources.

## Demo
![rakkess demo](doc/demo-small.png "rakkess demo")

## Examples
Show access for all resources
- ... at cluster scope
  ```bash
  rakkess
  ```

- ... in some namespace
  ```bash
  rakkess --namespace default
  ```

- ... with verbs
  ```bash
  rakkess --verbs get,delete,watch,proxy
  ```

- ... for another user
  ```bash
  rakkess --as other-user
  ```

- ... and combine with common `kubectl` parameters
  ```bash
  KUBECONFIG=otherconfig rakkess --context other-context
  ```

Also see [Usage](doc/USAGE.md).

## Installation

### Via krew
```bash
kubectl krew install access-matrix
```

### Binaries
When using the binaries for installation, also have a look at [doc/USAGE](doc/USAGE.md).

#### Linux
```bash
curl -Lo rakkess https://github.com/corneliusweig/rakkess/releases/download/v0.1.0/rakkess-linux-amd64.gz &&
  gunzip rakkess-linux-amd64.gz && chmod +x rakkess-linux-amd64 && mv rakkess-linux-amd64 $GOPATH/bin/rakkess
```

#### OSX
```bash
curl -Lo rakkess https://github.com/corneliusweig/rakkess/releases/download/v0.1.0/rakkess-darwin-amd64.gz &&
  gunzip rakkess-darwin-amd64.gz && chmod +x rakkess-darwin-amd64 && mv rakkess-darwin-amd64 $GOPATH/bin/rakkess
```

#### Windows
[https://github.com/corneliusweig/rakkess/releases/download/v0.1.0/rakkess-windows-amd64.zip](https://github.com/corneliusweig/rakkess/releases/download/v0.1.0/rakkess-windows-amd64.zip)

### From source

#### Build on host

Requirements:
 - go 1.11 or newer
 - GNU make
 - git

Compiling:
```bash
export PLATFORMS=$(go env GOOS)
make all   # binaries will be placed in out/
```

#### Build in docker
Requirements:
 - docker

Compiling:
```bash
mkdir rakkess && chdir rakkess
curl -Lo Dockerfile https://raw.githubusercontent.com/corneliusweig/rakkess/master/Dockerfile
docker build . -t rakkess-builder
docker run --rm -v $PWD:/go/bin/ --env PLATFORMS=$(go env GOOS) rakkess
docker rmi rakkess-builder
```
Binaries will be placed in the current directory.
