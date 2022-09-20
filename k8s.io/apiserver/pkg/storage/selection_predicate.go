package storage

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

type AttrFunc func(obj runtime.Object) (labels.Set, fields.Set, error)

type FieldMutationFunc func(obj runtime.Object, fieldSet fields.Set) error

func DefaultClusterScopedAttr(obj runtime.Object) (labels.Set, fields.Set, error) {
	metadata, err := meta.Accessor(obj)
	if err != nil {
		return nil, nil, err
	}
	fieldSet := fields.Set{
		"metadata.name": metadata.GetName(),
	}

	return labels.Set(metadata.GetLabels()), fieldSet, nil
}

func DefaultNamespaceScopedAttr(obj runtime.Object) (labels.Set, fields.Set, error) {
	metadata, err := meta.Accessor(obj)
	if err != nil {
		return nil, nil, err
	}
	fieldSet := fields.Set{
		"metadata.name":      metadata.GetName(),
		"metadata.namespace": metadata.GetNamespace(),
	}

	return labels.Set(metadata.GetLabels()), fieldSet, nil
}

func (f AttrFunc) WithFieldMutation(fieldMutator FieldMutationFunc) AttrFunc {
	return func(obj runtime.Object) (labels.Set, fields.Set, error) {
		labelSet, fieldSet, err := f(obj)
		if err != nil {
			return nil, nil, err
		}
		if err := fieldMutator(obj, fieldSet); err != nil {
			return nil, nil, err
		}
		return labelSet, fieldSet, nil
	}
}

type SelectionPredicate struct {
	Label               labels.Selector
	Field               fields.Selector
	GetAttrs            AttrFunc
	IndexLabels         []string
	IndexFields         []string
	Limit               int64
	Continue            string
	AllowWatchBookmarks bool
}

func (s *SelectionPredicate) Matches(obj runtime.Object) (bool, error) {
	if s.Empty() {
		return true, nil
	}
	labels, fields, err := s.GetAttrs(obj)
	if err != nil {
		return false, err
	}
	matched := s.Label.Matches(labels)
	if matched && s.Field != nil {
		matched = matched && s.Field.Matches(fields)
	}
	return matched, nil
}

func (s *SelectionPredicate) MatchesObjectAttributes(l labels.Set, f fields.Set) bool {
	if s.Label.Empty() && s.Field.Empty() {
		return true
	}
	matched := s.Label.Matches(l)
	if matched && s.Field != nil {
		matched = (matched && s.Field.Matches(f))
	}
	return matched
}

func (s *SelectionPredicate) MatchesSingle() (string, bool) {
	if len(s.Continue) > 0 {
		return "", false
	}
	// TODO: should be namespace.name
	if name, ok := s.Field.RequiresExactMatch("metadata.name"); ok {
		return name, true
	}
	return "", false
}

func (s *SelectionPredicate) Empty() bool {
	return s.Label.Empty() && s.Field.Empty()
}

func (s *SelectionPredicate) MatcherIndex() []MatchValue {
	var result []MatchValue
	for _, field := range s.IndexFields {
		if value, ok := s.Field.RequiresExactMatch(field); ok {
			result = append(result, MatchValue{IndexName: FieldIndex(field), Value: value})
		}
	}
	for _, label := range s.IndexLabels {
		if value, ok := s.Label.RequiresExactMatch(label); ok {
			result = append(result, MatchValue{IndexName: LabelIndex(label), Value: value})
		}
	}
	return result
}

func LabelIndex(label string) string {
	return "l:" + label
}

func FieldIndex(field string) string {
	return "f:" + field
}
