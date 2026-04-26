//go:generate go run cmd/cimgen/main.go
//go:generate go run cmd/cimgen/main.go -lang proto
//go:generate sh -c "protoc --proto_path=proto/definitions --go_out=proto/definitions --go_opt=paths=source_relative proto/definitions/*.proto"
package cimgo
