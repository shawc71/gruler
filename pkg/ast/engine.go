package ast

import (
	"gruler/pkg/proto"
	"sync"
)

type Engine struct {
	mutex   *sync.RWMutex
	program *Program
}

func NewEngine(program *Program) *Engine {
	return &Engine{
		mutex:   &sync.RWMutex{},
		program: program,
	}
}
func (e *Engine) Execute(request *proto.HttpRequest) ([]*proto.Action, error)  {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	actions := make([]*proto.Action, 0)
	for _, rule := range e.program.Rules {
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