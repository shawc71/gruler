package ast

import (
	"fmt"
	"gruler/pkg/proto"
	"strings"
)

// Maps request fields referenced in the rules json to the values of the field in the underlying
// request protobuf
type RequestFieldResolver struct {
	request *proto.HttpRequest
}

func NewRequestFieldResolver(request *proto.HttpRequest) *RequestFieldResolver {
	return &RequestFieldResolver{request: request}
}

func (f *RequestFieldResolver) GetFieldValue(fieldName string) (string, error) {
	arg := strings.Split(fieldName, ".")
	if arg[2] == "method" {
		return f.request.Method, nil
	} else if arg[2] == "header" {
		if len(arg) != 4 {
			return "", fmt.Errorf("invalid target %s", fieldName)
		}
		return f.request.Headers[arg[3]], nil
	} else if arg[2] == "clientIp" {
		return f.request.ClientIp, nil
	}
	return "", fmt.Errorf("invalid target %s", fieldName)
}
