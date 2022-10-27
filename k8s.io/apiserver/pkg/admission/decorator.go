package admission

type Decorator interface {
	Decorate(handler Interface, name string) Interface
}

type DecoratorFunc func(handler Interface, name string) Interface

func (d DecoratorFunc) Decorate(handler Interface, name string) Interface {
	return d(handler, name)
}

type Decorators []Decorator

func (d Decorators) Decorate(handler Interface, name string) Interface {
	result := handler
	for _, d := range d {
		result = d.Decorate(result, name)
	}

	return result
}
