package main

import (
	"encoding/binary"
	"flag"
	"github.com/golang/protobuf/proto"
	"gruler/pkg/ast"
	"gruler/pkg/conf_parser"
	localProto "gruler/pkg/proto"
	"gruler/pkg/throttle"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const sockAddr = "/tmp/gruler.sock"

type config struct {
	rulesFilePath string
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	conf := readConfig()
	program, err := conf_parser.NewConfParser(conf.rulesFilePath, throttle.NewBucketManager()).Read()
	if err != nil {
		log.Fatal(err)
	}
	engine := ast.NewEngine(program)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP)
	go signalHandler(signals, engine, conf)

	listener := startServer()
	log.Println("Started server...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}
		go handleConnection(conn, engine)
	}

}

func signalHandler(signals chan os.Signal, engine *ast.Engine, conf *config) {
	for {
		sig := <-signals
		if sig != syscall.SIGHUP {
			log.Println("Unknown Signal ", sig)
			continue
		}
		program, err := conf_parser.NewConfParser(conf.rulesFilePath, throttle.NewBucketManager()).Read()
		if err != nil {
			log.Println("Unable to parse rules: ", err)
			continue
		}
		engine.UpdateProgram(program)
	}
}

func startServer() net.Listener {
	if err := os.RemoveAll(sockAddr); err != nil {
		log.Fatal(err)
	}

	listener, err := net.Listen("unix", sockAddr)
	if err != nil {
		log.Fatal("unable to start listener:", err)
	}
	return listener
}

func handleConnection(conn net.Conn, rulesEngine *ast.Engine) {
	defer conn.Close()
	for {
		request, err := readRequest(conn)
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			return
		}

		response := applyRules(rulesEngine, request)

		if err = writeResponse(conn, response); err != nil {
			log.Println(err)
			return
		}
	}
}

func writeResponse(conn net.Conn, response *localProto.Response) error {
	wireResponse, err := proto.Marshal(response)
	if err != nil {
		return err
	}

	if err = writeResponseSize(conn, wireResponse); err != nil {
		return err
	}

	if _, err = conn.Write(wireResponse); err != nil {
		return err
	}
	return nil
}

func writeResponseSize(conn net.Conn, wireResponse []byte) error {
	responseSizeBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(responseSizeBuf, uint32(len(wireResponse)))
	if _, err := conn.Write(responseSizeBuf); err != nil {
		return err
	}
	return nil
}

func applyRules(rulesEngine *ast.Engine, request *localProto.Request) *localProto.Response {
	startTime := time.Now()
	requestFieldResolver := ast.NewRequestFieldResolver(request.GetHttpRequest())
	actions, err := rulesEngine.Execute(requestFieldResolver)
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
	}
	delta := time.Since(startTime).Nanoseconds()
	response := &localProto.Response{}
	if err == nil {
		response = &localProto.Response{
			ExecutionTime: delta,
			Actions:       actions,
			Success:       true,
		}
	} else {
		response = &localProto.Response{
			ExecutionTime: delta,
			Success:       false,
		}
	}
	return response
}

func readRequest(conn net.Conn) (*localProto.Request, error) {
	sizeBuf := make([]byte, 4)
	if _, err := conn.Read(sizeBuf); err != nil {
		return nil, err
	}

	messageSize := binary.BigEndian.Uint32(sizeBuf)
	dataBuf := make([]byte, messageSize)

	if _, err := conn.Read(dataBuf); err != nil {
		return nil, err
	}
	request := localProto.Request{}
	if err := proto.Unmarshal(dataBuf, &request); err != nil {
		return nil, err
	}
	return &request, nil
}

func readConfig() *config {
	conf := config{}
	flag.StringVar(&conf.rulesFilePath, "rulesFile", "", "path to the file containing the rules")

	flag.Parse()

	if conf.rulesFilePath == "" {
		flag.Usage()
		os.Exit(1)
	}

	return &conf
}
