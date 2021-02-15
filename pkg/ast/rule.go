package ast

import (
	"fmt"
	"gruler/pkg/proto"
	"gruler/pkg/throttle"
	"strings"
	"time"
)

type Rule struct {
	RuleId    string
	Condition Condition
	Action    Action
}

type ThrottleConfig struct {
	MaxAmount     int64
	RefillTime    time.Duration
	RefillAmount  int64
	EachUnique    string
	BucketManager *throttle.BucketManager
}

type Action struct {
	ActionType   string
	Status       int64
	HeaderName   string
	HeaderVal    string
	ThrottleConf ThrottleConfig
}

func (r *Rule) Evaluate(request *proto.HttpRequest) (*proto.Action, bool, error) {
	isTrue, err := r.Condition.True(request)
	if err != nil {
		return nil, false, err
	}
	if !isTrue {
		return &proto.Action{}, false, nil
	}
	if r.Action.ActionType == "throttle" {
		key, err := r.generateThrottleKey(request)
		if err != nil {
			return nil, false, err
		}
		throttleConf := r.Action.ThrottleConf
		bucket := throttleConf.BucketManager.GetBucket(key, throttleConf.MaxAmount, throttleConf.RefillTime, throttleConf.RefillAmount)
		shouldThrottle := bucket.GetThrottleDecision()
		if !shouldThrottle {
			// throttle rules only apply if the condition evaluates to true and throttle decision is true
			return nil, false, nil
		}
	}
	return &proto.Action{
		ActionType:  r.Action.ActionType,
		RuleId:      r.RuleId,
		Status:      r.Action.Status,
		HeaderName:  r.Action.HeaderName,
		HeaderValue: r.Action.HeaderVal,
	}, true, nil
}

func (r *Rule) generateThrottleKey(request *proto.HttpRequest) (string, error) {
	if r.Action.ThrottleConf.EachUnique == "" {
		return r.RuleId, nil
	}
	eachUniqueVal := r.Action.ThrottleConf.EachUnique
	arg := strings.Split(r.Action.ThrottleConf.EachUnique, ".")
	if arg[2] == "method" {
		return fmt.Sprintf("%s.%s", r.RuleId, request.Method), nil
	} else if arg[2] == "header" {
		if len(arg) != 4 {
			return "", fmt.Errorf("invalid EachUnique %s", eachUniqueVal)
		}
		return fmt.Sprintf("%s.%s", r.RuleId, request.Headers[arg[3]]), nil
	} else if arg[2] == "clientIp" {
		return fmt.Sprintf("%s.%s", request.ClientIp), nil
	}
	return "", fmt.Errorf("invalid EachUnique %s", eachUniqueVal)
}
