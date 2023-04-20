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
<!-- <img src="https://img.shields.io/github/v/release/tobias-mayer/vector-db" alt="drawing"/>
<img src="https://img.shields.io/docker/pulls/tobias-mayer/vector-db" alt="drawing"/>
<img src="https://img.shields.io/github/downloads/tobias-mayer/vector-db/total.svg" alt="drawing"/> -->
</div>

# Table of Contents
<!--ts-->
- [vector-db](#vector-db)
- [Table of Contents](#table-of-contents)
- [Examples](#examples)
    - [Hello World](#hello-world)
    - [HTTP Transport](#http-transport)
- [Makefile Targets](#makefile-targets)

<!--te-->

# Examples

### Hello World
```sh
$> go run examples/helloworld/helloworld.go
Output:
The following vectors are the closest neighbors based on cosine similarity:
vector: [0.16 0.9], distance: 0.997870
vector: [0.014 0.99], distance: 0.995346
vector: [0.01 0.88], distance: 0.995074
vector: [0.009 0.95], distance: 0.994885
vector: [0 0.91], distance: 0.993884
```

### HTTP Transport
```sh
$> go run examples/db/http.go
$> curl -d '{"vector":[0.1, 0.9], "numberOfNeighbors":3}' -H "Content-Type: application/json" -X POST http://localhost:8080/search
Output:
{"vectors":[[0.16,0.9],[0.014,0.99],[0.01,0.88]],"distances":[0.9978698605247033,0.9953458731615661,0.9950743923646557]}
```



# Makefile Targets
```sh
$> make
bootstrap                      install build deps
build                          build golang binary
clean                          clean up environment
cover                          display test coverage
fmt                            format go files
help                           list makefile targets
install                        install golang binary
lint                           lint go files
pre-commit                     run pre-commit hooks
run                            run the app
test                           display test coverage
```
