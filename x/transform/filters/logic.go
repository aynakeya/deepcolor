package filters

import "github.com/aynakeya/deepcolor/transform"

type OrFilter struct {
	transform.BaseFilter
	Filters []transform.Filter
}

func Or(filters []transform.Filter) transform.Filter {
	return &OrFilter{
		BaseFilter: transform.BaseFilter{
			Type: "or",
		},
		Filters: filters,
	}
}

func (o *OrFilter) Check(value interface{}) bool {
	for _, filter := range o.Filters {
		if filter.Check(value) {
			return true
		}
	}
	return false
}

type AndFilter struct {
	transform.BaseFilter
	Filters []transform.Filter
}

func And(filters []transform.Filter) transform.Filter {
	return &AndFilter{
		BaseFilter: transform.BaseFilter{
			Type: "and",
		},
		Filters: filters,
	}
}

func (a *AndFilter) Check(value interface{}) bool {
	for _, filter := range a.Filters {
		if !filter.Check(value) {
			return false
		}
	}
	return true
}

type NotFilter struct {
	transform.BaseFilter
	Filter transform.Filter
}

func Not(filter transform.Filter) transform.Filter {
	return &NotFilter{
		BaseFilter: transform.BaseFilter{
			Type: "not",
		},
		Filter: filter,
	}
}

func (n *NotFilter) Check(value interface{}) bool {
	return !n.Filter.Check(value)
}
