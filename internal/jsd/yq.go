package jsd

import (
	yq "github.com/mikefarah/yq/v4/pkg/yqlib"
	"gopkg.in/op/go-logging.v1"
)

func init() {
	yq.GetLogger().SetBackend(&backend{})
}

type backend struct {
}

func (b *backend) Log(logging.Level, int, *logging.Record) (err error) {
	return
}

func (b *backend) GetLevel(string) (s logging.Level) {
	return
}
func (b *backend) SetLevel(logging.Level, string) {
}

func (b *backend) IsEnabledFor(logging.Level, string) (en bool) {
	return
}
