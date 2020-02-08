# Contributing
`rakkess` uses GitHub to manage reviews of pull requests.

* If you have a trivial fix or improvement, go ahead and create a pull request.

* Code must be properly formatted using `go fmt`

* If you plan to do something more involved, first raise an issue to discuss
  your idea. This will avoid unnecessary work.

* Relevant coding style guidelines are  the [Go Code Review Comments](https://code.google.com/p/go-wiki/wiki/CodeReviewComments)
  and the _Formatting and style_ section of Peter Bourgon's [Go: Best Practices for Production Environments](http://peter.bourgon.org/go-in-production/#formatting-and-style).

## Building & Testing

* Build via `make dev` to create the binary file `./rakkess`.
* Run unit tests with: `make test`
* Run coverage with: `make coverage`

## Pull Request Checklist

* Use the [latest stable Go release](https://golang.org/dl/)

* Branch from master and, if needed, rebase to the current master branch before submitting your pull request.
  If it doesn't merge cleanly with master you will be asked to rebase your changes.

* Commits should be small units of work with one topic. Each commit should be correct independently.

* Add tests relevant to the fixed bug or new feature.

* Add a [DCO](https://developercertificate.org/) / `Signed-off-by` line in any commit message (`git commit --signoff`).

### Setup of your local environment

- `go env` shows helpful info about the current env setup for go.
- Check [here](https://github.com/golang/go/wiki/GOPATH) for more info on setting `$GOPATH` and `$GOROOT` env vars.

#### Quick Start:

1. `mkdir -p $HOME/go/src`
2. `export GOPATH=$HOME/go`
3. `go get -u github.com/corneliusweig/rakkess`
4. Set `$GOROOT` depending on your OS and Go installation method:
   - MacOS, Go installed via brew: `export GOROOT=/usr/local/opt/go/libexec/`
5. Now you should be able to build:
   - `cd $GOPATH/src/github.com/corneliusweig/rakkess/`
   - `make dev`

## Releases

This is a checklist for new releases:

0. Create release notes in `doc/releases` by running `hack/release_notes.sh`.
0. Update usage instructions, if applicable.
0. Create a new tag via `hack/make_tag.sh`.
0. Push the tag to GitHub `git push origin v1.3.1`.
0. Check that PR is auto-created in [krew-index](https://github.com/kubernetes-sigs/krew-index).
0. When the release is created, update its summary with the release notes.
