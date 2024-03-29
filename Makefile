all: proto server driver

server:
	go build -o ./dist/gruler_server ./cmd/server

driver:
	go build -o ./dist/test_driver ./cmd/test_driver

proto:
	protoc --go_out=paths=source_relative:. -I. pkg/proto/*.proto
	protoc --java_out=java-proto/src/main/java -I. pkg/proto/*.proto

fmt:
	gofmt -s -w pkg cmd

clean:
	rm -rf ./dist
	rm -rf ./java-proto/src/main/java/*
	rm ./pkg/proto/*.pb.go