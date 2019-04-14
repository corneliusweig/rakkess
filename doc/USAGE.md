<!-- DO NOT MOVE THIS FILE, BECAUSE IT NEEDS A PERMANENT ADDRESS -->

# Usage
![rakkess demo](demo-small.png "rakkess demo")

If you installed via [krew](https://github.com/kubernetes-sigs/krew) do
```bash
kubectl access-matrix
```

## Options

- `--verbs` show access for given verbs (valid verbs are `get`, `list`, `watch`, `create`, `update`, `delete`, `proxy`).

- `--namespace` show access rights for the given namespace. Also restricts the list to namespaced resources.

- `--verbosity` set the log level (one of debug, info, warn, error, fatal, panic).

## Examples
Show access for all resources
- ... at cluster scope
  ```bash
  kubectl access-matrix
  ```
  This defaults to the verbs `list`, `create`, `update`, and `delete` because they are the most common ones.

- ... in some namespace
  ```bash
  kubectl access-matrix --namespace default
  ```

- ... with verbs
  ```bash
  kubectl access-matrix --verbs get,delete,watch,proxy
  ```

- ... for another user
  ```bash
  kubectl access-matrix --as other-user
  ```

- ... and combine with common `kubectl` parameters
  ```bash
  KUBECONFIG=otherconfig kubectl access-matrix --context other-context

## Getting help
```bash
kubectl access-matrix help
```
Note that in the help, the tool is referred to as `rakkess`, which is the standard name when installed as stand-alone tool.

## Completion
Completion does currently not work when used as a `kubectl` plugin. When used stand-alone, you can do
```bash
source <(rakkess completion bash) # for bash users
source <(rakkess completion zsh)  # for zsh users
```
Also see `rakkess completion --help` for further instructions.

## Installation

### Via krew
If you do not have `krew` installed, visit [https://github.com/kubernetes-sigs/krew](https://github.com/kubernetes-sigs/krew).
```bash
kubectl krew install access-matrix
```

### As `kubectl` plugin
Most users will have installed `rakkess` via [krew](https://github.com/kubernetes-sigs/krew),
so the plugin is already correctly installed.
Otherwise, rename `rakkess` to `kubectl-access_matrix` and put it in some directory from your `$PATH` variable.
Then you can invoke the plugin via `kubectl access-matrix`

### Standalone
Put the `rakkess` binary in some directory from your `$PATH` variable. For example
```bash
sudo mv -i rakkess /usr/bin/rakkess
```
Then you can invoke the plugin via `rakkess`
