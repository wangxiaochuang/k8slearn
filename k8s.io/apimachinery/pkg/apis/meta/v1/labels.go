package v1

func CloneSelectorAndAddLabel(selector *LabelSelector, labelKey, labelValue string) *LabelSelector {
	if labelKey == "" {
		return selector
	}

	newSelector := selector.DeepCopy()

	if newSelector.MatchLabels == nil {
		newSelector.MatchLabels = make(map[string]string)
	}

	newSelector.MatchLabels[labelKey] = labelValue

	return newSelector
}

func AddLabelToSelector(selector *LabelSelector, labelKey, labelValue string) *LabelSelector {
	if labelKey == "" {
		// Don't need to add a label.
		return selector
	}
	if selector.MatchLabels == nil {
		selector.MatchLabels = make(map[string]string)
	}
	selector.MatchLabels[labelKey] = labelValue
	return selector
}

func SelectorHasLabel(selector *LabelSelector, labelKey string) bool {
	return len(selector.MatchLabels[labelKey]) > 0
}
