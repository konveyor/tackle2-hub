package proxy

import (
	"fmt"
	"testing"
)

func TestProxyGetUpdate(t *testing.T) {
	for _, id := range []uint{1, 2} { // Existing proxies in Hub have IDs 1, 2
		t.Run(fmt.Sprint(id), func(t *testing.T) {
			// Get.
			orig, err := Proxy.Get(id)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Update.
			update := orig
			update.Host = "127.0.0.1"
			update.Port = 8081
			update.Enabled = true
			err = Proxy.Update(update)
			if err != nil {
				t.Errorf(err.Error())
			}

			updated, err := Proxy.Get(update.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if updated.Host != update.Host || updated.Port != update.Port || updated.Enabled != update.Enabled {
				t.Errorf("Different response error. Got %+v, expected %+v", updated, update)
			}

			// Update back to original.
			err = Proxy.Update(orig)
			if err != nil {
				t.Fatalf(err.Error())
			}
		})
	}
}

func TestSeedProxyList(t *testing.T) {
	got, err := Proxy.List()
	if err != nil {
		t.Errorf(err.Error())
	}
	foundHttp, foundHttps := false, false
	for _, r := range got {
		if r.Kind == "http" {
			foundHttp = true
		}
		if r.Kind == "https" {
			foundHttps = true
		}
	}
	if !foundHttp {
		t.Errorf("Cannot find HTTP proxy.")
	}
	if !foundHttps {
		t.Errorf("Cannot find HTTPS proxy.")
	}
}
