package settings

import (
	"os"
	"strconv"
)

const (
	EnvShutdownTimeout = "SHUTDOWN_TIMEOUT"
	EnvRecoveryAddress = "RECOVERY_ADDRESS"
)

type Lifecycle struct {
	Shutdown struct {
		Timeout int
	}
	Recovery struct {
		Address string
	}
}

func (r *Lifecycle) Load() (err error) {
	s, found := os.LookupEnv(EnvShutdownTimeout)
	if found {
		n, _ := strconv.Atoi(s)
		r.Shutdown.Timeout = n
	} else {
		r.Shutdown.Timeout = 30 // 30 seconds.
	}
	r.Recovery.Address, found = os.LookupEnv(EnvRecoveryAddress)
	if !found {
		r.Recovery.Address = ":8083"
	}
	return
}
