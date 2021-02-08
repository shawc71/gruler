package main

import (
	"fmt"
	"gruler/pkg/ast"
	"gruler/pkg/conf_parser"
	"gruler/pkg/proto"
	"os"
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
	response, err := engine.Execute(httpRequest)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("%+v", response)
}
