package profile

import k8s "sigs.k8s.io/controller-runtime/pkg/client"

type Resolver interface {
	Match(capability string) (names []string, err error)
}

type BaseResolver struct {
	client k8s.Client
}

type AddonResolver struct {
	BaseResolver
}

func (r *AddonResolver) Match(capability string) (names []string, err error) {
	return
}

type ComponentResolver struct {
	BaseResolver
}

func (r *ComponentResolver) Match(capability string) (names []string, err error) {
	return
}
