.PHONY: all generate proto

all: generate proto

generate:
	go generate ./...

proto:
	protoc --go_out=. --proto_path=proto/definitions proto/definitions/*.proto
