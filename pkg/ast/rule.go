package ast

import (
	"fmt"
	"gruler/pkg/proto"
	"gruler/pkg/throttle"
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

func (r *Rule) Evaluate(requestFieldResolver *RequestFieldResolver) (*proto.Action, bool, error) {
	isTrue, err := r.Condition.True(requestFieldResolver)
	if err != nil {
		return nil, false, err
	}
	if !isTrue {
		return &proto.Action{}, false, nil
	}
	if r.Action.ActionType == "throttle" {
		key, err := r.generateThrottleKey(requestFieldResolver)
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

func (r *Rule) generateThrottleKey(requestFieldResolver *RequestFieldResolver) (string, error) {
	if r.Action.ThrottleConf.EachUnique == "" {
		return r.RuleId, nil
	}
	fieldVal, err := requestFieldResolver.GetFieldValue(r.Action.ThrottleConf.EachUnique)
	if err != nil {
		return "", fmt.Errorf("invalid EachUnique %s", r.Action.ThrottleConf.EachUnique)
	}
	return fmt.Sprintf("%s.%s", r.RuleId, fieldVal), nil
}
