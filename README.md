# vector-db

<div align="center">
A simple vector database that can be used to search for similar vectors in logarithmic time.
<br>
<br>
<img src="https://github.com/tobias-mayer/vector-db/actions/workflows/test.yml/badge.svg" alt="drawing"/>
<img src="https://github.com/tobias-mayer/vector-db/actions/workflows/lint.yml/badge.svg" alt="drawing"/>
<a href="https://codecov.io/gh/tobias-mayer/vector-db" >
<img src="https://codecov.io/gh/tobias-mayer/vector-db/branch/master/graph/badge.svg?token=V3XINHNCKM"/>
</a>
<img src="https://img.shields.io/github/v/release/tobias-mayer/vector-db" alt="drawing"/>
<img src="https://img.shields.io/docker/pulls/tobias-mayer/vector-db" alt="drawing"/>
<img src="https://img.shields.io/github/downloads/tobias-mayer/vector-db/total.svg" alt="drawing"/>
</div>

# Table of Contents
<!--ts-->
- [vector-db](#vector-db)
- [Table of Contents](#table-of-contents)
- [Demo Application](#demo-application)
- [Makefile Targets](#makefile-targets)

<!--te-->

# Demo Application

```sh
$> vector-db -h
vector-db

Usage:
  vector-db [flags]
  vector-db [command]

Available Commands:
  start       Start the vector-db
  help        Help about any command
  version     vector-db version

Flags:
  -h, --help   help for vector-db

Use "vector-db [command] --help" for more information about a command.
```

```sh
$> vector-db start
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
