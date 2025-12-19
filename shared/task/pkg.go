package task

import "strings"

// States
// also used as events:
// - Postponed
// - QuotaBlocked
const (
	Created      = "Created"
	Ready        = "Ready"
	Postponed    = "Postponed"
	QuotaBlocked = "QuotaBlocked"
	Pending      = "Pending"
	Running      = "Running"
	Succeeded    = "Succeeded"
	Failed       = "Failed"
	Canceled     = "Canceled"
)

// Events
const (
	AddonSelected    = "AddonSelected"
	ExtSelected      = "ExtensionSelected"
	ImageError       = "ImageError"
	PodNotFound      = "PodNotFound"
	PodCreated       = "PodCreated"
	PodPending       = "PodPending"
	PodUnschedulable = "PodUnschedulable"
	PodRunning       = "PodRunning"
	PodSucceeded     = "PodSucceeded"
	PodFailed        = "PodFailed"
	PodDeleted       = "PodDeleted"
	Escalated        = "Escalated"
	Released         = "Released"
	ContainerKilled  = "ContainerKilled"
)

// Group Modes
const (
	Batch    = "Batch"
	Pipeline = "Pipeline"
)

// ExtEnv returns an environment variable named namespaced to an extension.
// Format: _EXT_<extension_<var>.
func ExtEnv(extension string, envar string) (s string) {
	s = strings.Join(
		[]string{
			"_EXT",
			strings.ToUpper(extension),
			envar,
		},
		"_")
	return
}
