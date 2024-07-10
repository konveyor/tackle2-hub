package api

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/jortel/go-utils/logr"
)

const (
	Unit = time.Second
)

type Server struct {
	Handler http.Handler
	Address string
	Name    string
}

func (r *Server) Run(ctx context.Context, wg *sync.WaitGroup) {
	Log = logr.WithName(r.Name)
	srv := &http.Server{
		Addr:    r.Address,
		Handler: r.Handler,
	}
	go func() {
		Log.Info("api server starting", "address", r.Address)
		err := srv.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			Log.Error(err, "api server failed", "address", r.Address)
		}
	}()

	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			force, cancel := context.WithTimeout(context.Background(), r.timeout())
			defer cancel()
			err := srv.Shutdown(force)
			if err != nil {
				Log.Error(err, "api server shutdown failed", "address", r.Address)
			}
			return
		}
	}()
}

func (r *Server) timeout() (d time.Duration) {
	d = Unit * time.Duration(Settings.Shutdown.Timeout)
	return
}
