package v1alpha1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	Ready = meta.Condition{
		Type:   "Ready",
		Status: meta.ConditionTrue,
	}
	ImageNotDefined = meta.Condition{
		Type:    "Ready",
		Status:  meta.ConditionTrue,
		Reason:  "Validation failed.",
		Message: "Either image or container.image must be specified.",
	}
)
