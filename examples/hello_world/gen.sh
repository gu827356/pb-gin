#! /bin/bash
set -e

# build bin
go build -o temp/protoc-gen-gin ../../cmd/protoc-gen-gin

protoc -I ../../protos -I . --go_out . \
  --go_opt=module=github.com/gu827356/pb-gin/examples/hello_world hello_world.proto

protoc -I ../../protos -I . --plugin=./temp/protoc-gen-gin --gin_out . \
  --gin_opt=module=github.com/gu827356/pb-gin/examples/hello_world hello_world.proto
