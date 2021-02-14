package main

import (
	"encoding/binary"
	"fmt"
	"github.com/golang/protobuf/proto"
	localProto "gruler/pkg/proto"
	"log"
	"net"
	"time"
)

func main() {
	conn, _ := net.Dial("unix", "/tmp/gruler.sock")

	startTime := time.Now()
	writeDummyRequest(conn)
	delta := time.Since(startTime).Nanoseconds()

	response := readResponse(conn)
	fmt.Printf("%v\n Took %v\n", response, delta)
}

func writeDummyRequest(conn net.Conn) {
	headers := make(map[string]string)
	headers["host"] = "example.com"
	httpRequest := &localProto.HttpRequest{
		Method:   "PUT",
		ClientIp: "127.0.0.1",
		Headers:  headers,
	}
	request := &localProto.Request{HttpRequest: httpRequest}
	data, err := proto.Marshal(request)
	size := len(data)
	sizeBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBuf, uint32(size))
	_, err = conn.Write(sizeBuf)
	if err != nil {
		log.Fatal(err)
	}
	_, err = conn.Write(data)
	if err != nil {
		log.Fatal(err)
	}
}

func readResponse(conn net.Conn) *localProto.Response {
	sizeBuf := make([]byte, 4)
	if _, err := conn.Read(sizeBuf); err != nil {
		log.Fatal(err)
	}

	messageSize := binary.BigEndian.Uint32(sizeBuf)
	dataBuf := make([]byte, messageSize)

	if _, err := conn.Read(dataBuf); err != nil {
		log.Fatal(err)
	}
	response := localProto.Response{}
	if err := proto.Unmarshal(dataBuf, &response); err != nil {
		log.Fatal(err)
	}
	return &response
}