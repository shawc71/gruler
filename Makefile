all: proto server

server:
	go build -o ./dist/gruler

proto:
	protoc --go_out=paths=source_relative:. -I. pkg/proto/*.proto

clean:
	rm -rf ./dist