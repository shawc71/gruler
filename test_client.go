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
	headers := make(map[string]string)
	headers["host"] = "example.com"
	httpRequest := &localProto.HttpRequest{
		Method:  "PUT",
		ClientIp: "127.0.0.1",
		Headers: headers,
	}
	request := &localProto.Request{HttpRequest: httpRequest}
	data, err := proto.Marshal(request)
	size := len(data)
	fmt.Println(size)
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
	time.Sleep(30 * time.Second)
}