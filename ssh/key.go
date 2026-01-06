/*
Package ssh provides a SSH related functionality.
*/
package ssh

import "strings"

// Key is and SSH key.
type Key struct {
	ID         uint
	Name       string
	Content    string
	Passphrase string
}

// Add self to the default agent.
func (k Key) Add() (err error) {
	err = agent.Add(k)
	return
}

// Formatted returns the formatted key.
func (k *Key) Formatted() (out string) {
	if k.Content != "" {
		out = strings.TrimSpace(k.Content) + "\n"
	}
	return
}
