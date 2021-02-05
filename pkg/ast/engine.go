package ast

import (
	"gruler/pkg/proto"
)

type Engine struct {
	Program *Program
}

func (e *Engine) Execute(request *proto.HttpRequest) ([]*proto.Action, error)  {
	actions := make([]*proto.Action, 0)
	for _, rule := range e.Program.Rules {
		action, isTrue, err := rule.Evaluate(request)
		if err != nil {
			return nil, err
		}
		if isTrue {
			actions = append(actions, action)
		}
	}
	return actions, nil
}