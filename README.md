# CIMgo

CIMgo processes CGMES/CIM XML/RDF files in Go and protobuf.

## How to Build and Run

Make sure that you have cloned the repo recursively to include the CGMES schema files from ENTSO-E

    git clone --recurse-submodules [...]

or clone the submodule in a second step

    git submodule update --init --recursive

Ensure that GOPATH is set and included in your PATH.

Use the `go run` command, specifying the target language with the `-lang` flag.

```bash
go run cmd/cimgen/main.go -lang proto
```

Alternatively, you can install cimgen.

    go install ./...

## How to Test

Run the test suite using the `go test` command. The `-v` flag provides verbose output.

    go test -v ./...


## proto generation

Ensure that the protobuf tools are installed

    protoc --version
    
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest # If you also need gRPC services

Generate
    
    protoc --go_out=. --proto_path=./proto/definitions  proto/definitions/*.proto
