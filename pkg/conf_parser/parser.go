package conf_parser

import (
	"encoding/json"
	"fmt"
	"gruler/pkg/ast"
	"io/ioutil"
)

func Read(path string) (*ast.Program, error) {
	json, err := jsonFromFile(path)
	if err != nil {
		return nil, err
	}
	rules, err := parseRules(json)
	if err != nil {
		return nil, err
	}
	return &ast.Program{
		Rules: rules,
	}, nil
}

func parseRules(json map[string]interface{}) ([]ast.Rule, error) {
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
		condition, err := parseCondition(jsonRule["condition"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		rule.Condition = condition

		action, err := parseAction(jsonRule["action"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		rule.Action = action
		ruleList = append(ruleList, rule)
	}
	return ruleList, nil
}

func parseAction(actionJson map[string]interface{}) (ast.Action, error) {
	if _, exists := actionJson["type"]; !exists {
		return ast.Action{}, fmt.Errorf("'type' is required on action %v", actionJson)
	}
	actionType := actionJson["type"].(string)
	if actionType == "setHeader" {
		return parseSetHeaderAction(actionJson)
	}
	if actionType  == "block" {
		return parseBlockAction(actionJson)
	}
	return ast.Action{}, fmt.Errorf("invalid action type '%v'", actionJson)
}

func parseBlockAction(actionJson map[string]interface{}) (ast.Action, error) {
	action := ast.Action{}
	action.ActionType = "block"
	action.Status = int64(actionJson["statusCode"].(float64))
	return action, nil
}

func parseSetHeaderAction(actionJson map[string]interface{}) (ast.Action, error) {
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

func parseCondition(conditionJson map[string]interface{}) (ast.Condition, error) {
	if len(conditionJson) != 1 {
		return nil, fmt.Errorf("invalid condition %v", conditionJson)
	}
	for conditionType, value := range conditionJson {
		if conditionType == "eq" {
			return parseEqCondition(conditionJson, conditionType, value)
		} else if conditionType == "and" {
			return parseAndCondition(value)
		} else if conditionType == "or" {
			return parseOrCondition(value)
		} else if conditionType == "not" {
			return parseNotCondition(value)
		}
		return nil, fmt.Errorf("invalid condition %v:%v", conditionType, value)
	}
	return nil, fmt.Errorf("invalid condition %v", conditionJson)
}

func parseOrCondition(conditionJson interface{}) (ast.Condition, error) {
	parsedNestedConditions, err := parseConditionList(conditionJson)
	if err != nil {
		return nil, err
	}
	return &ast.OrCondition{Conditions: parsedNestedConditions}, nil
}

func parseAndCondition(conditionJson interface{}) (ast.Condition, error) {
	parsedNestedConditions, err := parseConditionList(conditionJson)
	if err != nil {
		return nil, err
	}
	return &ast.AndCondition{Conditions: parsedNestedConditions}, nil
}

func parseConditionList(value interface{}) ([]ast.Condition, error) {
	nestedConditions := value.([]interface{})
	parsedNestedConditions := make([]ast.Condition, 0)
	for _, nestedCondition := range nestedConditions {
		conditionJson := nestedCondition.(map[string]interface{})
		parsedCondition, err := parseCondition(conditionJson)
		if err != nil {
			return nil, err
		}
		parsedNestedConditions = append(parsedNestedConditions, parsedCondition)
	}
	return parsedNestedConditions, nil
}

func parseEqCondition(conditionJson map[string]interface{}, key string, value interface{}) (ast.Condition, error) {
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

func parseNotCondition(value interface{}) (ast.Condition, error) {
	wrappedConditionJson := value.(map[string]interface{})
	wrappedCondition, err := parseCondition(wrappedConditionJson)
	if err != nil {
		return nil, err
	}
	return &ast.NotCondition{WrappedCondition: wrappedCondition}, nil
}

func jsonFromFile(path string) (map[string]interface{}, error) {
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
