package command

import (
	"time"

	"github.com/konveyor/tackle2-hub/shared/command"
)

type Buffer = command.Buffer

const (
	// Backoff rate increment.
	Backoff = time.Millisecond * 100
	// MaxBackoff max backoff.
	MaxBackoff = 10 * Backoff
	// MinBackoff minimum backoff.
	MinBackoff = Backoff
)

// Writer reports command output.
// Provides both io.Reader and io.Writer.
// Command output is buffered (rate-limited) and reported.
type Writer struct {
	Buffer
	reporter *Reporter
	backoff  time.Duration
	end      chan any
	ended    chan any
}

// Write command output.
func (w *Writer) Write(p []byte) (n int, err error) {
	n, err = w.Buffer.Write(p)
	if err != nil {
		return
	}
	switch w.reporter.Verbosity {
	case LiveOutput:
		if w.ended == nil {
			w.end = make(chan any)
			w.ended = make(chan any)
			go w.report()
		}
	}
	return
}

// End of writing.
func (w *Writer) End() {
	if w.end == nil {
		return
	}
	close(w.end)
	<-w.ended
	close(w.ended)
	w.end = nil
}

// Reporter returns the reporter.
func (w *Writer) Reporter() *Reporter {
	return w.reporter
}

// report in task Report.Activity.
// Rate limited.
func (w *Writer) report() {
	w.backoff = MinBackoff
	ended := false
	for {
		select {
		case <-w.end:
			ended = true
		case <-time.After(w.backoff):
		}
		n := w.reporter.Output(w.Bytes())
		w.adjustBackoff(n)
		if ended && n == 0 {
			break
		}
	}
	w.ended <- true
}

// adjustBackoff adjust the backoff as needed.
// incremented when output reported.
// decremented when no outstanding output reported.
func (w *Writer) adjustBackoff(reported int) {
	if reported > 0 {
		if w.backoff < MaxBackoff {
			w.backoff += Backoff
		}
	} else {
		if w.backoff > MinBackoff {
			w.backoff -= Backoff
		}
	}
}
