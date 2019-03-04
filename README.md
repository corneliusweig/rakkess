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
There are several ways to install `rakkess`. The recommended installation method is via `krew`.

### Via krew
Krew is a `kubectl` plugin manager. If you have not yet installed `krew`, get it at
[https://github.com/GoogleContainerTools/krew](https://github.com/GoogleContainerTools/krew).
Then installation is as simple as
```bash
kubectl krew install access-matrix
```
The plugin will be available as `kubectl access-matrix`, see [doc/USAGE](doc/USAGE.md) for further details.

### Binaries
When using the binaries for installation, also have a look at [doc/USAGE](doc/USAGE.md).

#### Linux
```bash
curl -Lo rakkess.gz https://github.com/corneliusweig/rakkess/releases/download/v0.2.0/rakkess-linux-amd64.gz && \
  gunzip rakkess.gz && chmod +x rakkess && mv rakkess $GOPATH/bin/
```

#### OSX
```bash
curl -Lo rakkess.gz https://github.com/corneliusweig/rakkess/releases/download/v0.2.0/rakkess-darwin-amd64.gz && \
  gunzip rakkess.gz && chmod +x rakkess && mv rakkess $GOPATH/bin/
```

#### Windows
[https://github.com/corneliusweig/rakkess/releases/download/v0.2.0/rakkess-windows-amd64.zip](https://github.com/corneliusweig/rakkess/releases/download/v0.2.0/rakkess-windows-amd64.zip)

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

## Users

| What are others saying about rakkess? |
| ---- |
| _“Well, that looks handy! `rakkess`, a kubectl plugin to show an access matrix for all available resources.”_ – [@mhausenblas](https://twitter.com/mhausenblas/status/1100673166303739905) |
| _“that's indeed pretty helpful. `rakkess --as system:serviceaccount:my-ns:my-sa -n my-ns` prints the access matrix of a service account in a namespace”_ – [@fakod](https://twitter.com/fakod/status/1100764745957658626) |
| _“THE BOMB. Love it.”_ – [@ralph_squillace](https://twitter.com/ralph_squillace/status/1100844255830896640) |
| _“This made my day. Well, not actually today but I definitively will use it a lot.”_ – [@Soukron](https://twitter.com/Soukron/status/1100690060129775617) |

