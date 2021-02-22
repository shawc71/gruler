package ast

type Condition interface {
	True(requestFieldResolver *RequestFieldResolver) (bool, error)
}

type EqCondition struct {
	Left  string
	Right string
}

func (e *EqCondition) True(requestFieldResolver *RequestFieldResolver) (bool, error) {
	fieldValue, err := requestFieldResolver.GetFieldValue(e.Left)
	if err != nil {
		return false, err
	}
	return fieldValue == e.Right, nil
}

type AndCondition struct {
	Conditions []Condition
}

func (a *AndCondition) True(requestFieldResolver *RequestFieldResolver) (bool, error) {
	condResult := true
	for _, cond := range a.Conditions {
		isTrue, err := cond.True(requestFieldResolver)
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

func (o *OrCondition) True(requestFieldResolver *RequestFieldResolver) (bool, error) {
	condResult := false
	for _, cond := range o.Conditions {
		isTrue, err := cond.True(requestFieldResolver)
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

func (n *NotCondition) True(requestFieldResolver *RequestFieldResolver) (bool, error) {
	isTrue, err := n.WrappedCondition.True(requestFieldResolver)
	if err != nil {
		return false, err
	}
	return !isTrue, nil
}

type InCondition struct {
	Left  string
	Right map[string]bool
}

func (e *InCondition) True(requestFieldResolver *RequestFieldResolver) (bool, error) {
	fieldValue, err := requestFieldResolver.GetFieldValue(e.Left)
	if err != nil {
		return false, err
	}
	_, exists := e.Right[fieldValue]
	return exists, nil
}
