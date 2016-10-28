package store

import "math"

type Aggregation interface {
	Accumulate(_json interface{})
	Compute() float64
}

type Aggregations struct {
	context
	Aggregations map[string]Aggregation
}

type SumAggregation struct {
	context
	path *node32
	sum  float64
}

func (a *SumAggregation) Accumulate(_json interface{}) {
	if value, valid := a.getFloat64(a.path, _json); valid {
		a.sum += value
	}
}

func (a *SumAggregation) Compute() float64 {
	return a.sum
}

type CountAggregation struct {
	context
	path  *node32
	count uint64
}

func (a *CountAggregation) Accumulate(_json interface{}) {
	if a.getValid(a.path, _json) {
		a.count++
	}
}

func (a *CountAggregation) Compute() float64 {
	return float64(a.count)
}

type CountAllAggregation struct {
	count uint64
}

func (a *CountAllAggregation) Accumulate(_json interface{}) {
	a.count++
}

func (a *CountAllAggregation) Compute() float64 {
	return float64(a.count)
}

type AvgAggregation struct {
	context
	path  *node32
	sum   float64
	count uint64
}

func (a *AvgAggregation) Accumulate(_json interface{}) {
	if value, valid := a.getFloat64(a.path, _json); valid {
		a.sum += value
		a.count++
	}
}

func (a *AvgAggregation) Compute() float64 {
	if a.count == 0 {
		return 0
	}
	return a.sum / float64(a.count)
}

type StdDevAggregation struct {
	context
	path            *node32
	sum, sumSquared float64
	count           uint64
}

func (a *StdDevAggregation) Accumulate(_json interface{}) {
	if value, valid := a.getFloat64(a.path, _json); valid {
		a.sum += value
		a.sumSquared += value * value
		a.count++
	}
}

func (a *StdDevAggregation) Compute() float64 {
	if a.count == 0 {
		return 0
	}
	count := float64(a.count)
	avg := a.sum / count
	return math.Sqrt((a.sumSquared / count) - avg*avg)
}

type MinAggregation struct {
	context
	path *node32
	min  float64
}

func (a *MinAggregation) Accumulate(_json interface{}) {
	if value, valid := a.getFloat64(a.path, _json); valid && value < a.min {
		a.min = value
	}
}

func (a *MinAggregation) Compute() float64 {
	return a.min
}

type MaxAggregation struct {
	context
	path *node32
	max  float64
}

func (a *MaxAggregation) Accumulate(_json interface{}) {
	if value, valid := a.getFloat64(a.path, _json); valid && value > a.max {
		a.max = value
	}
}

func (a *MaxAggregation) Compute() float64 {
	return a.max
}

func (a *Aggregations) Process(node *node32) {
	for node != nil {
		switch node.pegRule {
		case rulee:
			a.Process(node.up)
		case ruleaggregate:
			a.ProcessAggregate(node.up)
		}
		node = node.next
	}
	return
}

func (a *Aggregations) ProcessAggregate(node *node32) {
	for node != nil {
		if node.pegRule == ruleaggregate_clause {
			a.ProcessAggregateClause(node.up)
		}
		node = node.next
	}
	return
}

func (a *Aggregations) ProcessAggregateClause(node *node32) {
	function, typ := "", ruleUnknown
	var selector *node32
	for node != nil {
		switch node.pegRule {
		case rulefunction:
			function = a.Node(node)
		case ruleselector:
			typ, selector = a.ProcessSelector(node.up)
		case ruleword:
			name := a.Node(node)
			switch function {
			case "sum":
				a.Aggregations[name] = &SumAggregation{
					context: a.context,
					path:    selector,
				}
			case "count":
				switch typ {
				case rulepath:
					a.Aggregations[name] = &CountAggregation{
						context: a.context,
						path:    selector,
					}
				case rulewildcard:
					a.Aggregations[name] = &CountAllAggregation{}
				}
			case "avg":
				a.Aggregations[name] = &AvgAggregation{
					context: a.context,
					path:    selector,
				}
			case "stddev":
				a.Aggregations[name] = &StdDevAggregation{
					context: a.context,
					path:    selector,
				}
			case "min":
				a.Aggregations[name] = &MinAggregation{
					context: a.context,
					path:    selector,
					min:     math.MaxFloat64,
				}
			case "max":
				a.Aggregations[name] = &MaxAggregation{
					context: a.context,
					path:    selector,
				}
			}
		}
		node = node.next
	}
	return
}

func (a *Aggregations) ProcessSelector(node *node32) (typ pegRule, value *node32) {
	for node != nil {
		switch node.pegRule {
		case rulepath:
			return rulepath, node.up
		case rulewildcard:
			return rulewildcard, node.up
		}
		node = node.next
	}
	return
}
