all: proto server driver

server:
	go build -o ./dist/gruler_server ./cmd/server

driver:
	go build -o ./dist/test_driver ./cmd/test_driver

proto:
	protoc --go_out=paths=source_relative:. -I. pkg/proto/*.proto

clean:
	rm -rf ./dist
