package ast

import (
	"gruler/pkg/proto"
	"log"
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

func (e *Engine) Execute(requestFieldResolver *RequestFieldResolver) ([]*proto.Action, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	actions := make([]*proto.Action, 0)
	for _, rule := range e.program.Rules {
		action, isTrue, err := rule.Evaluate(requestFieldResolver)
		if err != nil {
			return nil, err
		}
		if isTrue {
			actions = append(actions, action)
		}
	}
	return actions, nil
}

func (e *Engine) UpdateProgram(program *Program) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	log.Print("updated program")
	e.program = program
}
