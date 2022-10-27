package admission

import "context"

func newReinvocationHandler(admissionChain Interface) Interface {
	return &reinvoker{admissionChain}
}

type reinvoker struct {
	admissionChain Interface
}

func (r *reinvoker) Admit(ctx context.Context, a Attributes, o ObjectInterfaces) error {
	panic("not implemented")
}

func (r *reinvoker) Validate(ctx context.Context, a Attributes, o ObjectInterfaces) error {
	if validator, ok := r.admissionChain.(ValidationInterface); ok {
		return validator.Validate(ctx, a, o)
	}
	return nil
}

func (r *reinvoker) Handles(operation Operation) bool {
	return r.admissionChain.Handles(operation)
}
