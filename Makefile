all: proto server driver

server:
	go build -o ./dist/server ./cmd/server

driver:
	go build -o ./dist/driver ./cmd/test_driver

proto:
	protoc --go_out=paths=source_relative:. -I. pkg/proto/*.proto

clean:
	rm -rf ./dist