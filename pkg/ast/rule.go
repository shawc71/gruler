package ast

import (
	"gruler/pkg/proto"
)

type Rule struct {
	RuleId    string
	Condition Condition
	Action    Action
}

type Action struct {
	ActionType string
	Status     int64
	HeaderName string
	HeaderVal  string
}

func (r *Rule) Evaluate(request *proto.HttpRequest) (*proto.Action, bool, error) {
	isTrue, err := r.Condition.True(request)
	if err != nil {
		return nil, false, err
	}
	if isTrue {
		return &proto.Action{
			ActionType:  r.Action.ActionType,
			RuleId:      r.RuleId,
			Status:      r.Action.Status,
			HeaderName:  r.Action.HeaderName,
			HeaderValue: r.Action.HeaderVal,
		}, true, nil
	}
	return &proto.Action{}, false, nil
}
