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

	count := 1000
	startTime := time.Now()
	response := &localProto.Response{}
	conn, _ := net.Dial("unix", "/tmp/gruler.sock")
	for i := 0; i < count; i++ {
		writeDummyRequest(conn, i)
		response = readResponse(conn)
	}
	delta := time.Since(startTime).Milliseconds()
	fmt.Printf("%v\n Took %v\n", response, float64(delta) / float64(count))
}

func writeDummyRequest(conn net.Conn, i int) {
	headers := make(map[string]string)
	headers["host"] = "example.com"
	httpRequest := &localProto.HttpRequest{
		Method:   "PUT",
		ClientIp: fmt.Sprintf("127.0.0.%d", i),
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