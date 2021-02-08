package main

import (
	"fmt"
	"gruler/pkg/ast"
	"gruler/pkg/conf_parser"
	"gruler/pkg/proto"
	"os"
	"time"
)

func main() {
	headers := make(map[string]string)
	headers["host"] = "example.com"
	httpRequest := &proto.HttpRequest{
		Method:  "PUT",
		ClientIp: "127.0.0.1",
		Headers: headers,
	}
	_ = proto.Request{HttpRequest: httpRequest}

	prog, err := conf_parser.Read("rules.json")
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}
	engine := &ast.Engine{
		Program:  prog,
	}
	startTime := time.Now()
	response, err := engine.Execute(httpRequest)
	delta := time.Since(startTime).Nanoseconds()
	if err != nil {
		fmt.Print(err)
	}
	resp := proto.Response{
		ExecutionTime: delta,
		Actions:       response,
	}
	fmt.Printf("%+v", resp)
}
