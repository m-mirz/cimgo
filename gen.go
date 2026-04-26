//go:generate go run cmd/cimgen/main.go
//go:generate go run cmd/cimgen/main.go -lang proto
//go:generate sh -c "protoc --go_out=. --proto_path=./proto/definitions  proto/definitions/*.proto"
package cimgo
