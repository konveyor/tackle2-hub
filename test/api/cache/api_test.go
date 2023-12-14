package cache

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
)

func TestCacheGet(t *testing.T) {
	cache := api.Cache{}
	err := RichClient.Client.Get(api.CacheRoot, &cache)
	if err != nil {
		t.Errorf(err.Error())
	}
	// Cache could be disabled, so just hit the cache endpoint and check the request has not failed
}