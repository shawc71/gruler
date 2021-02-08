package ast

import "C"
import (
	"fmt"
	"gruler/pkg/proto"
	"strings"
)

type Condition interface {
	True(request *proto.HttpRequest) (bool, error)
}

type EqCondition struct {
	Left  string
	Right string
}

func (e *EqCondition) True(request *proto.HttpRequest) (bool, error) {
	arg := strings.Split(e.Left, ".")
	if arg[2] == "method" {
		return request.Method == e.Right, nil
	} else if arg[2] == "header" {
		if len(arg) != 4 {
			return false, fmt.Errorf("invalid target %s", e.Left)
		}
		return request.Headers[arg[3]] == e.Right, nil
	} else if arg[2] == "clientIp" {
		return request.ClientIp == e.Right, nil
	}
	return false, fmt.Errorf("invalid target %s", e.Left)
}

type AndCondition struct {
	Conditions []Condition
}

func (a *AndCondition) True(request *proto.HttpRequest) (bool, error) {
	condResult := true
	for _, cond := range a.Conditions {
		isTrue, err := cond.True(request)
		if err != nil {
			return false, err
		}
		condResult = condResult && isTrue
		if !condResult {
			return condResult, nil
		}
	}
	return condResult, nil
}
type OrCondition struct {
	Conditions []Condition
}

func (o *OrCondition) True(request *proto.HttpRequest) (bool, error) {
	condResult := false
	for _, cond := range o.Conditions {
		isTrue, err := cond.True(request)
		if err != nil {
			return false, err
		}
		condResult = condResult || isTrue
		if condResult {
			return condResult, nil
		}
	}
	return condResult, nil
}

type NotCondition struct {
	WrappedCondition Condition
}

func (n *NotCondition) True(request *proto.HttpRequest) (bool, error) {
	isTrue, err := n.WrappedCondition.True(request)
	if err != nil {
		return false, err
	}
	return !isTrue, nil
}
