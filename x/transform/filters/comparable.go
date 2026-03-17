package filters

import (
	"github.com/aynakeya/deepcolor/transform"
)

//type ComparableOperator string
//
//const (
//	CmpEqual            ComparableOperator = "eq"
//	CmpNotEqual                            = "ne"
//	CmpLessThan                            = "lt"
//	CmpGreaterThan                         = "gt"
//	CmpLessThanEqual                       = "lte"
//	CmpGreaterThanEqual                    = "gte"
//)
//
//type ComparableFilter[T comparable] struct {
//	transform.BaseFilter
//	Value    T
//	Operator ComparableOperator
//}
//
//func NewComparableFilter[T comparable](value T, operator ComparableOperator) transform.Filter {
//	return &ComparableFilter[T]{
//		BaseFilter: transform.BaseFilter{
//			Type: "ComparableFilter",
//		},
//		Value:    value,
//		Operator: operator,
//	}
//}
//
//func (c *ComparableFilter[T]) Check(value interface{}) bool {
//	val, ok := value.(T)
//	if !ok {
//		return false
//	}
//
//	switch c.Operator {
//	case CmpEqual:
//		return val == c.Value
//	case CmpNotEqual:
//		return val != c.Value
//	case CmpLessThan:
//		return val < c.Value
//	case CmpGreaterThan:
//		return val > c.Value
//	case CmpLessThanEqual:
//		return val <= c.Value
//	case CmpGreaterThanEqual:
//		return val >= c.Value
//	default:
//		return false
//	}
//}

type EqualFilter[T comparable] struct {
	transform.BaseFilter
	Negate bool
	Value  T
}

func Equal[T comparable](value T) transform.Filter {
	return &EqualFilter[T]{
		BaseFilter: transform.BaseFilter{
			Type: "equal",
		},
		Negate: false,
		Value:  value,
	}
}

func NotEqual[T comparable](value T) transform.Filter {
	return &EqualFilter[T]{
		BaseFilter: transform.BaseFilter{
			Type: "equal",
		},
		Negate: true,
		Value:  value,
	}
}

func (f *EqualFilter[T]) Check(value interface{}) bool {
	val, ok := value.(T)
	if !ok {
		return false
	}
	return (val == f.Value) != f.Negate
}

type InFilter[T comparable] struct {
	transform.BaseFilter
	Negate bool
	Values []T
}

func In[T comparable](values []T) transform.Filter {
	return &InFilter[T]{
		BaseFilter: transform.BaseFilter{
			Type: "in",
		},
		Negate: false,
		Values: values,
	}
}

func NotIn[T comparable](values []T) transform.Filter {
	return &InFilter[T]{
		BaseFilter: transform.BaseFilter{
			Type: "in",
		},
		Negate: true,
		Values: values,
	}
}

func (f *InFilter[T]) Check(value interface{}) bool {
	val, ok := value.(T)
	if !ok {
		return false
	}
	for _, v := range f.Values {
		if val == v {
			return !f.Negate
		}
	}
	return f.Negate
}
