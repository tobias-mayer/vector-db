# vector-db

<div align="center">
A general purpose project template for golang CLI applications
<br>
<br>
This template serves as a starting point for golang commandline applications it is based on golang projects that I consider high quality and various other useful blog posts that helped me understanding golang better.
<br>
<br>
<img src="https://github.com/tobias-mayer/vector-db/actions/workflows/test.yml/badge.svg" alt="drawing"/>
<img src="https://github.com/tobias-mayer/vector-db/actions/workflows/lint.yml/badge.svg" alt="drawing"/>
<img src="https://pkg.go.dev/badge/github.com/tobias-mayer/vector-db.svg" alt="drawing"/>
<img src="https://codecov.io/gh/tobias-mayer/vector-db/branch/main/graph/badge.svg" alt="drawing"/>
<img src="https://img.shields.io/github/v/release/tobias-mayer/vector-db" alt="drawing"/>
<img src="https://img.shields.io/docker/pulls/tobias-mayer/vector-db" alt="drawing"/>
<img src="https://img.shields.io/github/downloads/tobias-mayer/vector-db/total.svg" alt="drawing"/>
</div>

# Table of Contents
<!--ts-->
   * [vector-db](#vector-db)
   * [Features](#features)
   * [Project Layout](#project-layout)
   * [How to use this template](#how-to-use-this-template)
   * [Demo Application](#demo-application)
   * [Makefile Targets](#makefile-targets)
   * [Contribute](#contribute)

<!-- Added by: morelly_t1, at: Tue 10 Aug 2021 08:54:24 AM CEST -->

<!--te-->

# Features
- [goreleaser](https://goreleaser.com/) with `deb.` and `.rpm` packer and container (`docker.hub` and `ghcr.io`) releasing including `manpages` and `shell completions` and grouped Changelog generation.
- [golangci-lint](https://golangci-lint.run/) for linting and formatting
- [Github Actions](.github/worflows) Stages (Lint, Test (`windows`, `linux`, `mac-os`), Build, Release) 
- [Gitlab CI](.gitlab-ci.yml) Configuration (Lint, Test, Build, Release)
- [cobra](https://cobra.dev/) example setup including tests
- [Makefile](Makefile) - with various useful targets and documentation (see Makefile Targets)
- [Github Pages](_config.yml) using [jekyll-theme-minimal](https://github.com/pages-themes/minimal) (checkout [https://tobias-mayer.github.io/vector-db/](https://tobias-mayer.github.io/vector-db/))
- Useful `README.md` badges
- [pre-commit-hooks](https://pre-commit.com/) for formatting and validating code before committing

# Project Layout
* [assets/](https://pkg.go.dev/github.com/tobias-mayer/vector-db/assets) => docs, images, etc
* [cmd/](https://pkg.go.dev/github.com/tobias-mayer/vector-db/cmd)  => commandline configurartions (flags, subcommands)
* [pkg/](https://pkg.go.dev/github.com/tobias-mayer/vector-db/pkg)  => packages that are okay to import for other projects
* [internal/](https://pkg.go.dev/github.com/tobias-mayer/vector-db/pkg)  => packages that are only for project internal purposes
- [`tools/`](tools/) => for automatically shipping all required dependencies when running `go get` (or `make bootstrap`) such as `golang-ci-lint` (see: https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module)
)
- [`scripts/`](scripts/) => build scripts 

# How to use this template
```sh
bash <(curl -s https://raw.githubusercontent.com/tobias-mayer/vector-db/master/install.sh)
```

In order to make the CI work you will need to have the following Secrets in your repository defined:

Repository  -> Settings -> Secrets & variables -> `CODECOV_TOKEN`, `DOCKERHUB_TOKEN` & `DOCKERHUB_USERNAME`

# Demo Application

```sh
$> vector-db -h
golang-cli project template demo application

Usage:
  vector-db [flags]
  vector-db [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  example     example subcommand which adds or multiplies two given integers
  help        Help about any command
  version     vector-db version

Flags:
  -h, --help   help for vector-db

Use "vector-db [command] --help" for more information about a command.
```

```sh
$> vector-db example 2 5 --add
7

$> vector-db example 2 5 --multiply
10
```

# Makefile Targets
```sh
$> make
bootstrap                      install build deps
build                          build golang binary
clean                          clean up environment
cover                          display test coverage
docker-build                   dockerize golang application
fmt                            format go files
help                           list makefile targets
install                        install golang binary
lint                           lint go files
pre-commit                     run pre-commit hooks
run                            run the app
test                           display test coverage
```

# Contribute
If you find issues in that setup or have some nice features / improvements, I would welcome an issue or a PR :)
