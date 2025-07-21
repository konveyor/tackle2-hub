package jsd

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func New(client client.Client) (m *Manager) {
	m = &Manager{Client: client}
	return
}
