package main

import (
	"encoding/binary"
	"fmt"
	"github.com/golang/protobuf/proto"
	localProto "gruler/pkg/proto"
	"log"
	"net"
	"sync"
	"time"
)

func main() {

	conn, _ := net.Dial("unix", "/tmp/gruler.sock")
	waitGroup := &sync.WaitGroup{}
	count := 20
	//startTime := time.Now()
	for {
		for i := 0; i < count; i++ {
			waitGroup.Add(1)
			response := callServer(conn, i, waitGroup)
			fmt.Printf("%v\n", response)
		}
		fmt.Println("----------------------------------------------")
		time.Sleep(time.Second * 2)
	}
	/*
		delta := time.Since(startTime).Milliseconds()
		fmt.Printf("Took %v\n", float64(delta)/float64(count))
	*/
}

func callServer(conn net.Conn, i int, waitGroup *sync.WaitGroup) *localProto.Response {
	defer waitGroup.Done()

	writeDummyRequest(conn, i)
	response := readResponse(conn)
	return response
}

func writeDummyRequest(conn net.Conn, i int) {
	headers := make(map[string]string)
	headers["host"] = "throttled.example.com"
	httpRequest := &localProto.HttpRequest{
		Headers: headers,
	}
	httpRequest.Method = "PUT"
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
