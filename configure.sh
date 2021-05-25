#!/usr/bin/env bash

[ $# -eq 1 ] && protoc $1/$1pb/$1.proto --go_out=plugins=grpc:. || echo "Usage: $0 <name>"
