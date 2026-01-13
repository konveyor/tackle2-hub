package heap

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/shared/settings"
)

var (
	Settings = &settings.Settings
	Log      = logr.New("heap", Settings.Log.Heap)
)

func Free() {
	runtime.GC()
	debug.FreeOSMemory()
	Print()
}

func Print() {
	mb := func(n uint64) (f float64) {
		f = float64(n) / 1024 / 1024
		return
	}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memory := m.HeapSys - m.HeapReleased
	reserved := m.HeapIdle - m.HeapReleased
	allocated := m.HeapAlloc
	unknown := memory - allocated - reserved
	s := "\nHEAP:\n"
	s += "_______________________\n"
	s += fmt.Sprintf("Memory   = %.2f MB\n", mb(memory))
	s += fmt.Sprintf("Used     = %.2f MB\n", mb(allocated))
	s += fmt.Sprintf("Reserved = %.2f MB\n", mb(reserved))
	s += fmt.Sprintf("Unknown  = %.2f MB\n", mb(unknown))
	Log.V(1).Info(s)
}

func Monitor() {
	delay := time.Duration(Settings.Frequency.Heap) * time.Second
	if delay == 0 {
		return // disabled
	}
	go func() {
		for {
			time.Sleep(delay)
			Free()
		}
	}()
}
