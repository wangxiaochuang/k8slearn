package errors

import (
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"
)

type MessageCountMap map[string]int

type Aggregate interface {
	error
	Errors() []error
	Is(error) bool
}

func NewAggregate(errlist []error) Aggregate {
	if len(errlist) == 0 {
		return nil
	}
	var errs []error
	for _, e := range errlist {
		if e != nil {
			errs = append(errs, e)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return aggregate(errs)
}

type aggregate []error

func (agg aggregate) Error() string {
	if len(agg) == 0 {
		return ""
	}
	if len(agg) == 1 {
		return agg[0].Error()
	}
	seenerrs := sets.NewString()
	result := ""
	agg.visit(func(err error) bool {
		msg := err.Error()
		if seenerrs.Has(msg) {
			return false
		}
		seenerrs.Insert(msg)
		if len(seenerrs) > 1 {
			result += ", "
		}
		result += msg
		return false
	})
	if len(seenerrs) == 1 {
		return result
	}
	return "[" + result + "]"
}

func (agg aggregate) Is(target error) bool {
	return agg.visit(func(err error) bool {
		return errors.Is(err, target)
	})
}

func (agg aggregate) visit(f func(err error) bool) bool {
	for _, err := range agg {
		switch err := err.(type) {
		case aggregate:
			if match := err.visit(f); match {
				return match
			}
		case Aggregate:
			for _, nestedErr := range err.Errors() {
				if match := f(nestedErr); match {
					return match
				}
			}
		default:
			if match := f(err); match {
				return match
			}
		}
	}

	return false
}

func (agg aggregate) Errors() []error {
	return []error(agg)
}

type Matcher func(error) bool

// 将满足fns条件的error剔除
func FilterOut(err error, fns ...Matcher) error {
	if err == nil {
		return nil
	}
	if agg, ok := err.(Aggregate); ok {
		return NewAggregate(filterErrors(agg.Errors(), fns...))
	}
	if !matchesError(err, fns...) {
		return err
	}
	return nil
}

func matchesError(err error, fns ...Matcher) bool {
	for _, fn := range fns {
		if fn(err) {
			return true
		}
	}
	return false
}

func filterErrors(list []error, fns ...Matcher) []error {
	result := []error{}
	for _, err := range list {
		r := FilterOut(err, fns...)
		if r != nil {
			result = append(result, r)
		}
	}
	return result
}

func Flatten(agg Aggregate) Aggregate {
	result := []error{}
	if agg == nil {
		return nil
	}
	for _, err := range agg.Errors() {
		if a, ok := err.(Aggregate); ok {
			r := Flatten(a)
			if r != nil {
				result = append(result, r.Errors()...)
			}
		} else {
			if err != nil {
				result = append(result, err)
			}
		}
	}
	return NewAggregate(result)
}

func CreateAggregateFromMessageCountMap(m MessageCountMap) Aggregate {
	if m == nil {
		return nil
	}
	result := make([]error, 0, len(m))
	for errStr, count := range m {
		var countStr string
		if count > 1 {
			countStr = fmt.Sprintf(" (repeated %v times)", count)
		}
		result = append(result, fmt.Errorf("%v%v", errStr, countStr))
	}
	return NewAggregate(result)
}

// 如果只有一个，就返回第一个
func Reduce(err error) error {
	if agg, ok := err.(Aggregate); ok && err != nil {
		switch len(agg.Errors()) {
		case 1:
			return agg.Errors()[0]
		case 0:
			return nil
		}
	}
	return err
}

func AggregateGoroutines(funcs ...func() error) Aggregate {
	errChan := make(chan error, len(funcs))
	for _, f := range funcs {
		go func(f func() error) { errChan <- f() }(f)
	}
	errs := make([]error, 0)
	for i := 0; i < cap(errChan); i++ {
		if err := <-errChan; err != nil {
			errs = append(errs, err)
		}
	}
	return NewAggregate(errs)
}

var ErrPreconditionViolated = errors.New("precondition is violated")
