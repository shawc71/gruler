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
	waitGroup.Add(1)
	response := callServer(conn, "127.0.0.0", waitGroup)
	fmt.Printf("%v\n", response)
	//throttleTest(conn)
	/*
		delta := time.Since(startTime).Milliseconds()
		fmt.Printf("Took %v\n", float64(delta)/float64(count))
	*/
}

func throttleTest(conn net.Conn) {
	waitGroup := &sync.WaitGroup{}
	count := 20
	//startTime := time.Now()
	for {
		for i := 0; i < count; i++ {
			waitGroup.Add(1)
			response := callServer(conn, "127.0.0.0", waitGroup)
			fmt.Printf("%v\n", response)
		}

		fmt.Println("Running for other ip")
		for i := 0; i < count; i++ {
			waitGroup.Add(1)
			response := callServer(conn, "127.0.0.1", waitGroup)
			fmt.Printf("%v\n", response)
		}
		fmt.Println("----------------------------------------------")
		time.Sleep(time.Second * 2)
	}
}

func callServer(conn net.Conn, clientIp string, waitGroup *sync.WaitGroup) *localProto.Response {
	defer waitGroup.Done()

	writeDummyRequest(conn, clientIp)
	response := readResponse(conn)
	return response
}

func writeDummyRequest(conn net.Conn, clientIp string) {
	headers := make(map[string]string)
	headers["host"] = "foo.example.com"
	httpRequest := &localProto.HttpRequest{
		Headers: headers,
	}
	httpRequest.ClientIp = clientIp
	httpRequest.Method = "OPTIONS"
	httpRequest.RequestUri = "/some-path/bar?userId=joe"
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
