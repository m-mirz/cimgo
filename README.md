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


## Architecture

The code generation process follows these main steps:

1.  **Schema Loading:** The tool begins by finding and parsing the relevant CIM RDF schema files based on the specified version and profiles.
2.  **Schema Processing:** It processes the parsed RDF data into an internal, language-agnostic representation of CIM classes, properties, datatypes, and their relationships.
3.  **Code Generation:** Using Go's `text/template` engine, it feeds the internal representation into language-specific templates (`lang-templates/*.tmpl`) to generate the final source code files.

## Key Go Files

*   `cmd/cimgen/main.go`: The main entry point for the CLI tool. It parses command-line arguments and orchestrates the code generation process.
*   `cim_generate.go`: Contains the core logic for driving the generation process for a specific language.
*   `cim_schema_import.go`: Handles the discovery and parsing of the CIM RDF schema files.
*   `cim_schema_processing.go`: Responsible for transforming the raw parsed schema into the internal representation used by the generators.
*   `generator_*.go`: A set of files (e.g., `generator_go.go`, `generator_cpp.go`) that implement the generation logic for each target language.
*   `templates.go`: Manages the embedded template files.
