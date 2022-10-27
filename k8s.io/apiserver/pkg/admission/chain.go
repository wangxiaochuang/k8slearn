package admission

import "context"

type chainAdmissionHandler []Interface

func NewChainHandler(handlers ...Interface) chainAdmissionHandler {
	return chainAdmissionHandler(handlers)
}

func (admissionHandler chainAdmissionHandler) Admit(ctx context.Context, a Attributes, o ObjectInterfaces) error {
	panic("not implemented")
}

func (admissionHandler chainAdmissionHandler) Validate(ctx context.Context, a Attributes, o ObjectInterfaces) error {
	for _, handler := range admissionHandler {
		if !handler.Handles(a.GetOperation()) {
			continue
		}
		if validator, ok := handler.(ValidationInterface); ok {
			err := validator.Validate(ctx, a, o)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (admissionHandler chainAdmissionHandler) Handles(operation Operation) bool {
	for _, handler := range admissionHandler {
		if handler.Handles(operation) {
			return true
		}
	}
	return false
}
