package main

import (
	"encoding/binary"
	"fmt"
	"github.com/golang/protobuf/proto"
	"gruler/pkg/ast"
	"gruler/pkg/conf_parser"
	localProto "gruler/pkg/proto"
	"log"
	"net"
	"os"
	"time"
)

const sockAddr = "/tmp/gruler.sock"

func main() {
	prog, err := conf_parser.Read("rules.json")
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}
	engine := &ast.Engine{
		Program:  prog,
	}

	if err = os.RemoveAll(sockAddr); err != nil {
		log.Fatal(err)
	}

	listener, err := net.Listen("unix", sockAddr)
	if err != nil {
		log.Fatal("unable to start listener:", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}
		go handleRequest(conn, engine)
	}

}

func handleRequest(conn net.Conn, rulesEngine *ast.Engine) {
	for {
		sizeBuf := make([]byte, 4)
		_, err := conn.Read(sizeBuf)
		if err != nil {
			log.Print(err)
			_ = conn.Close()
			return
		}
		messageSize := binary.BigEndian.Uint32(sizeBuf)
		dataBuf := make([]byte, messageSize)
		_, err = conn.Read(dataBuf)
		if err != nil {
			log.Print(err)
			_ = conn.Close()
			return
		}
		request := localProto.Request{}
		err = proto.Unmarshal(dataBuf, &request)
		if err != nil {
			log.Print(err)
			_ = conn.Close()
			return
		}

		startTime := time.Now()
		response, err := rulesEngine.Execute(request.GetHttpRequest())
		delta := time.Since(startTime).Nanoseconds()
		resp := &localProto.Response{}
		if err != nil {
			resp = &localProto.Response{
				ExecutionTime: delta,
				Success: false,
			}
		} else {
			resp = &localProto.Response{
				ExecutionTime: delta,
				Actions:       response,
				Success: true,
			}
		}

		wireResponse, err := proto.Marshal(resp)
		if err != nil {
			log.Print(err)
			_ = conn.Close()
			return
		}
		responseSizeBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(responseSizeBuf, uint32(len(wireResponse)))
		_, err = conn.Write(responseSizeBuf)
		if err != nil {
			log.Print(err)
			_ = conn.Close()
			return
		}
		_, err = conn.Write(wireResponse)
		if err != nil {
			log.Print(err)
			_ = conn.Close()
			return
		}
	}
}
