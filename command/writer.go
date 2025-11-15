package command

import (
	"fmt"
	"io"
)

// Buffer buffered command output writer.
type Buffer struct {
	buffer []byte
	read   int
}

// Write command output.
func (w *Buffer) Write(p []byte) (n int, err error) {
	n = len(p)
	w.buffer = append(w.buffer, p...)
	return
}

// Read the buffer.
func (w *Buffer) Read(p []byte) (n int, err error) {
	if w.read >= len(w.buffer) {
		err = io.EOF
		return
	}
	n = copy(p, w.buffer[w.read:])
	w.read += n
	return
}

// Seek to read position.
func (w *Buffer) Seek(offset int64, whence int) (n int64, err error) {
	switch whence {
	case io.SeekStart:
		n = offset
	case io.SeekCurrent:
		n = int64(w.read) + offset
	case io.SeekEnd:
		n = int64(len(w.buffer)) + offset
	default:
		err = fmt.Errorf("whence not valid: %d", whence)
		return
	}
	if n < 0 || n > int64(len(w.buffer)) {
		err = fmt.Errorf("out of bounds")
		return
	}
	w.read = int(n)
	return
}

// Bytes returns the buffer.
func (w *Buffer) Bytes() []byte {
	return w.buffer
}
