package conf_parser

import (
	"encoding/json"
	"fmt"
	"gruler/pkg/ast"
	"gruler/pkg/throttle"
	"io/ioutil"
	"strings"
	"time"
)

type confParser struct {
	filePath      string
	bucketManager *throttle.BucketManager
}

func NewConfParser(filePath string, bucketManager *throttle.BucketManager) *confParser {
	return &confParser{
		filePath:      filePath,
		bucketManager: bucketManager,
	}
}

func (c *confParser) Read() (*ast.Program, error) {
	rulesJson, err := c.jsonFromFile(c.filePath)
	if err != nil {
		return nil, err
	}
	rules, err := c.parseRules(rulesJson)
	if err != nil {
		return nil, err
	}
	return &ast.Program{
		Rules: rules,
	}, nil
}

func (c *confParser) parseRules(json map[string]interface{}) ([]ast.Rule, error) {
	_, exists := json["rules"]
	if !exists {
		return make([]ast.Rule, 0), nil
	}
	rules := json["rules"].([]interface{})
	rule := ast.Rule{}
	ruleList := make([]ast.Rule, 0)
	for _, j := range rules {
		jsonRule := j.(map[string]interface{})
		if _, exists := jsonRule["rule_id"]; !exists {
			return nil, fmt.Errorf("rule_id is missing in a rule")
		}
		rule.RuleId = jsonRule["rule_id"].(string)
		condition, err := c.parseCondition(jsonRule["condition"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		rule.Condition = condition

		action, err := c.parseAction(jsonRule["action"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		rule.Action = action
		ruleList = append(ruleList, rule)
	}
	return ruleList, nil
}

func (c *confParser) parseAction(actionJson map[string]interface{}) (ast.Action, error) {
	if _, exists := actionJson["type"]; !exists {
		return ast.Action{}, fmt.Errorf("'type' is required on action %v", actionJson)
	}
	actionType := actionJson["type"].(string)
	if actionType == "setHeader" {
		return c.parseSetHeaderAction(actionJson)
	}
	if actionType == "block" {
		return c.parseBlockAction(actionJson)
	}
	if actionType == "throttle" {
		return c.parseThrottleAction(actionJson)
	}
	return ast.Action{}, fmt.Errorf("invalid action type '%v'", actionJson)
}

func (c *confParser) parseBlockAction(actionJson map[string]interface{}) (ast.Action, error) {
	action := ast.Action{}
	action.ActionType = "block"
	action.Status = int64(actionJson["statusCode"].(float64))
	return action, nil
}

func (c *confParser) parseSetHeaderAction(actionJson map[string]interface{}) (ast.Action, error) {
	action := ast.Action{}
	action.ActionType = "setHeader"
	if _, exists := actionJson["headerName"]; !exists {
		return ast.Action{}, fmt.Errorf("'headerName' is required on action %v", actionJson)
	}
	if _, exists := actionJson["headerValue"]; !exists {
		return ast.Action{}, fmt.Errorf("'headerValue' is required on action %v", actionJson)
	}
	action.HeaderName = actionJson["headerName"].(string)
	action.HeaderVal = actionJson["headerValue"].(string)
	return action, nil
}

func (c *confParser) parseThrottleAction(actionJson map[string]interface{}) (ast.Action, error) {
	action := ast.Action{}
	action.ActionType = "throttle"
	throttleConf := ast.ThrottleConfig{}
	if _, exists := actionJson["max_tokens"]; !exists {
		return ast.Action{}, fmt.Errorf("'max_tokens' is required on action %v", actionJson)
	}
	if _, exists := actionJson["refill_amount"]; !exists {
		return ast.Action{}, fmt.Errorf("'refill_amount' is required on action %v", actionJson)
	}
	if _, exists := actionJson["refill_time"]; !exists {
		return ast.Action{}, fmt.Errorf("'refill_time' is required on action %v", actionJson)
	}
	if _, exists := actionJson["each_unique"]; exists {
		eachUniqueVal := actionJson["each_unique"].(string)
		throttleConf.EachUnique = strings.TrimSpace(eachUniqueVal)
	}
	throttleConf.MaxAmount = int64(actionJson["max_tokens"].(float64))
	throttleConf.RefillAmount = int64(actionJson["refill_amount"].(float64))
	throttleConf.RefillTime = time.Second * time.Duration(int64(actionJson["refill_time"].(float64)))
	throttleConf.BucketManager = c.bucketManager
	action.ThrottleConf = throttleConf
	return action, nil
}
func (c *confParser) parseCondition(conditionJson map[string]interface{}) (ast.Condition, error) {
	if len(conditionJson) != 1 {
		return nil, fmt.Errorf("invalid condition %v", conditionJson)
	}
	for conditionType, value := range conditionJson {
		switch conditionType {
		case "eq":
			return c.parseEqCondition(conditionJson, conditionType, value)
		case "and":
			return c.parseAndCondition(value)
		case "or":
			return c.parseOrCondition(value)
		case "not":
			return c.parseNotCondition(value)
		case "in":
			return c.parseInCondition(value)
		default:
			return nil, fmt.Errorf("invalid condition %v:%v", conditionType, value)
		}
	}
	return nil, fmt.Errorf("invalid condition %v", conditionJson)
}

func (c *confParser) parseOrCondition(conditionJson interface{}) (ast.Condition, error) {
	parsedNestedConditions, err := c.parseConditionList(conditionJson)
	if err != nil {
		return nil, err
	}
	return &ast.OrCondition{Conditions: parsedNestedConditions}, nil
}

func (c *confParser) parseAndCondition(conditionJson interface{}) (ast.Condition, error) {
	parsedNestedConditions, err := c.parseConditionList(conditionJson)
	if err != nil {
		return nil, err
	}
	return &ast.AndCondition{Conditions: parsedNestedConditions}, nil
}

func (c *confParser) parseConditionList(value interface{}) ([]ast.Condition, error) {
	nestedConditions := value.([]interface{})
	parsedNestedConditions := make([]ast.Condition, 0)
	for _, nestedCondition := range nestedConditions {
		conditionJson := nestedCondition.(map[string]interface{})
		parsedCondition, err := c.parseCondition(conditionJson)
		if err != nil {
			return nil, err
		}
		parsedNestedConditions = append(parsedNestedConditions, parsedCondition)
	}
	return parsedNestedConditions, nil
}

func (c *confParser) parseEqCondition(conditionJson map[string]interface{}, key string, value interface{}) (ast.Condition, error) {
	valueMap := value.(map[string]interface{})
	if len(valueMap) != 1 {
		return nil, fmt.Errorf("invalid condition %v", conditionJson)
	}
	for left, right := range valueMap {
		return &ast.EqCondition{
			Left:  left,
			Right: right.(string),
		}, nil
	}
	return nil, fmt.Errorf("invalid condition %v", conditionJson)
}

func (c *confParser) parseNotCondition(value interface{}) (ast.Condition, error) {
	wrappedConditionJson := value.(map[string]interface{})
	wrappedCondition, err := c.parseCondition(wrappedConditionJson)
	if err != nil {
		return nil, err
	}
	return &ast.NotCondition{WrappedCondition: wrappedCondition}, nil
}

func (c *confParser) parseInCondition(conditionJson interface{}) (ast.Condition, error) {
	valueMap := conditionJson.(map[string]interface{})
	if len(valueMap) != 1 {
		return nil, fmt.Errorf("invalid condition %v", conditionJson)
	}
	inSet := make(map[string]bool)
	for left, right := range valueMap {
		values := right.([]interface{})
		for _, item := range values {
			inSet[item.(string)] = true
		}
		return &ast.InCondition{
			Left:  left,
			Right: inSet,
		}, nil
	}
	return nil, fmt.Errorf("invalid condition %v", conditionJson)
}

func (c *confParser) jsonFromFile(path string) (map[string]interface{}, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = json.Unmarshal([]byte(data), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
