package ast

import (
	"fmt"
	"gruler/pkg/proto"
	"net/url"
	"strings"
)

// Maps request fields referenced in the rules json to the values of the field in the underlying
// request protobuf
type RequestFieldResolver struct {
	request           *proto.HttpRequest
	path              string
	rawQueryParams    string
	parsedQueryParams url.Values
}

func NewRequestFieldResolver(request *proto.HttpRequest) *RequestFieldResolver {
	vals := strings.SplitN(request.RequestUri, "?", 2)
	path := vals[0]
	rawQueryParams := ""
	if len(vals) > 1 {
		rawQueryParams = vals[1]
	}
	return &RequestFieldResolver{
		request:        request,
		path:           path,
		rawQueryParams: rawQueryParams,
	}
}

func (f *RequestFieldResolver) GetFieldValue(fieldName string) (string, error) {
	arg := strings.Split(fieldName, ".")
	switch arg[2] {
	case "method":
		return f.request.Method, nil
	case "header":
		return f.getHeaderVal(fieldName, arg)
	case "clientIp":
		return f.request.ClientIp, nil
	case "httpVersion":
		return f.request.HttpVersion, nil
	case "rawUri":
		return f.request.RequestUri, nil
	case "path":
		return f.path, nil
	case "rawQueryParams":
		return f.rawQueryParams, nil
	case "queryParam":
		return f.getQueryParam(fieldName, arg)
	default:
		return "", fmt.Errorf("invalid target %s", fieldName)
	}
}

func (f *RequestFieldResolver) getQueryParam(fieldName string, arg []string) (string, error) {
	if len(arg) != 4 {
		return "", fmt.Errorf("invalid target %s", fieldName)
	}
	if f.parsedQueryParams == nil {
		values, err := url.ParseQuery(f.rawQueryParams)
		if err != nil {
			return "", fmt.Errorf("unable to parse query string %s", f.rawQueryParams)
		}
		f.parsedQueryParams = values
	}
	return f.parsedQueryParams.Get(arg[3]), nil
}

func (f *RequestFieldResolver) getHeaderVal(fieldName string, arg []string) (string, error) {
	if len(arg) != 4 {
		return "", fmt.Errorf("invalid target %s", fieldName)
	}
	return f.request.Headers[arg[3]], nil
}
