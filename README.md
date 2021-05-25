# gRPC using protocol buffers

We will write three different subprojects that implements the different aspects of gRPC:
* Unary API
* Server streaming API
* Client streaming API
* Bi-Directional streaming API

The APIs are defined in a protocol buffer version 3 (```syntax = "proto3";```)

# Sub projects
* Greet
* Calculator
* Blog


## Sub project Greet
This project is completed in the lectures.

## Sub project Calculator
This sub project is implemented as homework.

## Sub project Blog
This is a more involved project using a CRUD-framework with MongoDB.

# go-code generation from the protocol buffers
We use a bash script ```configure.sh```
```
[ $# -eq 1 ]
    && protoc $1/$1pb/$1.proto --go_out=plugins=grpc:.
    || echo "Usage: $0 <name>"
```
which is called with 
```
./configure.sh [greet|calculator|blog]
```
in order to generate the go-code from the greet protocol buffer file.
we can replace ```greet``` with ```calculator``` to generate the calculator go-files. Once the ```blog``` project is up and running, this will be an option too.
