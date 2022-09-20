package conversion

import "k8s.io/apimachinery/third_party/forked/golang/reflect"

type Equalities struct {
	reflect.Equalities
}

func EqualitiesOrDie(funcs ...interface{}) Equalities {
	e := Equalities{reflect.Equalities{}}
	if err := e.AddFuncs(funcs...); err != nil {
		panic(err)
	}
	return e
}
