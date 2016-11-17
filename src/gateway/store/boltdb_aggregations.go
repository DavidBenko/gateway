package store

import (
	"errors"
	"math"
)

type Aggregation interface {
	Accumulate(_json interface{})
	Compute() interface{}
}

type Aggregations struct {
	context
	errors       []error
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

func (a *SumAggregation) Compute() interface{} {
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

func (a *CountAggregation) Compute() interface{} {
	return float64(a.count)
}

type CountAllAggregation struct {
	count uint64
}

func (a *CountAllAggregation) Accumulate(_json interface{}) {
	a.count++
}

func (a *CountAllAggregation) Compute() interface{} {
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

func (a *AvgAggregation) Compute() interface{} {
	if a.count == 0 {
		return 0
	}
	return a.sum / float64(a.count)
}

type VarAggregation struct {
	context
	path            *node32
	sum, sumSquared float64
	count           uint64
}

func (a *VarAggregation) Accumulate(_json interface{}) {
	if value, valid := a.getFloat64(a.path, _json); valid {
		a.sum += value
		a.sumSquared += value * value
		a.count++
	}
}

func (a *VarAggregation) Compute() interface{} {
	if a.count == 0 {
		return 0
	}
	count := float64(a.count)
	avg := a.sum / count
	return (a.sumSquared / count) - avg*avg
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

func (a *StdDevAggregation) Compute() interface{} {
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

func (a *MinAggregation) Compute() interface{} {
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

func (a *MaxAggregation) Compute() interface{} {
	return a.max
}

type CorrAggregation struct {
	context
	paths             []*node32
	xsum, xsumSquared float64
	ysum, ysumSquared float64
	xysum             float64
	count             uint64
}

func (a *CorrAggregation) Accumulate(_json interface{}) {
	y, yvalid := a.getFloat64(a.paths[0], _json)
	x, xvalid := a.getFloat64(a.paths[1], _json)
	if !(xvalid && yvalid) {
		return
	}
	a.xsum += x
	a.xsumSquared += x * x
	a.ysum += y
	a.ysumSquared += y * y
	a.xysum += x * y
	a.count++
}

func (a *CorrAggregation) Compute() interface{} {
	if a.count == 0 {
		return 0
	}
	count := float64(a.count)
	xavg, yavg := a.xsum/count, a.ysum/count
	xstddev := math.Sqrt((a.xsumSquared / count) - xavg*xavg)
	ystddev := math.Sqrt((a.ysumSquared / count) - yavg*yavg)
	if xstddev == 0 || ystddev == 0 {
		return 0
	}
	return (a.xysum/count - xavg*yavg) / (xstddev * ystddev)
}

type CovAggregation struct {
	context
	paths []*node32
	xsum  float64
	ysum  float64
	xysum float64
	count uint64
}

func (a *CovAggregation) Accumulate(_json interface{}) {
	y, yvalid := a.getFloat64(a.paths[0], _json)
	x, xvalid := a.getFloat64(a.paths[1], _json)
	if !(xvalid && yvalid) {
		return
	}
	a.xsum += x
	a.ysum += y
	a.xysum += x * y
	a.count++
}

func (a *CovAggregation) Compute() interface{} {
	if a.count == 0 {
		return 0
	}
	count := float64(a.count)
	xavg, yavg := a.xsum/count, a.ysum/count
	return a.xysum/count - xavg*yavg
}

type RegrAggregation struct {
	context
	paths        []*node32
	xsum, ysum   float64
	xxsum, xysum float64
	count        uint64
}

func (a *RegrAggregation) Accumulate(_json interface{}) {
	y, yvalid := a.getFloat64(a.paths[0], _json)
	x, xvalid := a.getFloat64(a.paths[1], _json)
	if !(xvalid && yvalid) {
		return
	}
	a.xsum += x
	a.ysum += y
	a.xxsum += x * x
	a.xysum += x * y
	a.count++
}

func (a *RegrAggregation) Compute() interface{} {
	result := make(map[string]interface{})
	result["a"] = 0
	result["b"] = 0
	if a.count == 0 {
		return result
	}
	count := float64(a.count)
	m := (count*a.xysum - a.xsum*a.ysum) / (count*a.xxsum - a.xsum*a.xsum)
	result["a"] = m
	result["b"] = (a.ysum / count) - (m*a.xsum)/count
	return result
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
	function := ""
	type Selector struct {
		typ      pegRule
		selector *node32
	}
	var selectors []Selector
	for node != nil {
		switch node.pegRule {
		case rulefunction:
			function = a.Node(node)
		case ruleselector:
			typ, selector := a.ProcessSelector(node.up)
			selectors = append(selectors, Selector{typ, selector})
		case ruleword:
			name := a.Node(node)
			switch function {
			case "sum":
				if len(selectors) != 1 {
					a.errors = append(a.errors, errors.New("sum takes 1 parameter"))
					return
				}
				a.Aggregations[name] = &SumAggregation{
					context: a.context,
					path:    selectors[0].selector,
				}
			case "count":
				if len(selectors) != 1 {
					a.errors = append(a.errors, errors.New("count takes 1 parameter"))
					return
				}
				switch selectors[0].typ {
				case rulepath:
					a.Aggregations[name] = &CountAggregation{
						context: a.context,
						path:    selectors[0].selector,
					}
				case rulewildcard:
					a.Aggregations[name] = &CountAllAggregation{}
				}
			case "avg":
				if len(selectors) != 1 {
					a.errors = append(a.errors, errors.New("avg takes 1 parameter"))
					return
				}
				a.Aggregations[name] = &AvgAggregation{
					context: a.context,
					path:    selectors[0].selector,
				}
			case "var":
				if len(selectors) != 1 {
					a.errors = append(a.errors, errors.New("var takes 1 parameter"))
					return
				}
				a.Aggregations[name] = &VarAggregation{
					context: a.context,
					path:    selectors[0].selector,
				}
			case "stddev":
				if len(selectors) != 1 {
					a.errors = append(a.errors, errors.New("stddev takes 1 parameter"))
					return
				}
				a.Aggregations[name] = &StdDevAggregation{
					context: a.context,
					path:    selectors[0].selector,
				}
			case "min":
				if len(selectors) != 1 {
					a.errors = append(a.errors, errors.New("min takes 1 parameter"))
					return
				}
				a.Aggregations[name] = &MinAggregation{
					context: a.context,
					path:    selectors[0].selector,
					min:     math.MaxFloat64,
				}
			case "max":
				if len(selectors) != 1 {
					a.errors = append(a.errors, errors.New("max takes 1 parameter"))
					return
				}
				a.Aggregations[name] = &MaxAggregation{
					context: a.context,
					path:    selectors[0].selector,
				}
			case "corr":
				if len(selectors) != 2 {
					a.errors = append(a.errors, errors.New("corr takes 2 parameters"))
					return
				}
				a.Aggregations[name] = &CorrAggregation{
					context: a.context,
					paths:   []*node32{selectors[0].selector, selectors[1].selector},
				}
			case "cov":
				if len(selectors) != 2 {
					a.errors = append(a.errors, errors.New("cov takes 2 parameters"))
					return
				}
				a.Aggregations[name] = &CovAggregation{
					context: a.context,
					paths:   []*node32{selectors[0].selector, selectors[1].selector},
				}
			case "regr":
				if len(selectors) != 2 {
					a.errors = append(a.errors, errors.New("cov takes 2 parameters"))
					return
				}
				a.Aggregations[name] = &RegrAggregation{
					context: a.context,
					paths:   []*node32{selectors[0].selector, selectors[1].selector},
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
