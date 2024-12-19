package task

import (
	"context"
	"io"
	"os"
	"sync"
	"time"

	liberr "github.com/jortel/go-utils/error"
	k8s2 "github.com/konveyor/tackle2-hub/k8s"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
	core "k8s.io/api/core/v1"
)

// LogCollector collect and report container logs.
type LogCollector struct {
	Owner     *LogManager
	Registry  map[string]*LogCollector
	DB        *gorm.DB
	Pod       *core.Pod
	Container *core.ContainerStatus
	//
	nBuf  int
	nSkip int64
}

// Begin - get container log and store in file.
// - Request logs.
// - Create file resource and attach to the task.
// - Register collector.
// - Write (copy) log.
// - Unregister collector.
func (r *LogCollector) Begin(task *Task, ctx context.Context) (err error) {
	reader, err := r.request()
	if err != nil {
		return
	}
	f, err := r.file(task)
	if err != nil {
		return
	}
	go func() {
		defer func() {
			_ = reader.Close()
			_ = f.Close()
			r.Owner.terminated(r)
		}()
		err := r.copy(reader, f, ctx)
		Log.Error(err, "")
	}()
	return
}

// key returns the collector key.
func (r *LogCollector) key() (key string) {
	key = r.Pod.Name + "." + r.Container.Name
	return
}

// request
func (r *LogCollector) request() (reader io.ReadCloser, err error) {
	options := &core.PodLogOptions{
		Container: r.Container.Name,
		Follow:    true,
	}
	clientSet, err := k8s2.NewClientSet()
	if err != nil {
		return
	}
	podClient := clientSet.CoreV1().Pods(Settings.Hub.Namespace)
	req := podClient.GetLogs(r.Pod.Name, options)
	reader, err = req.Stream(context.TODO())
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// name returns the canonical name for the container log.
func (r *LogCollector) name() (s string) {
	s = r.Container.Name + ".log"
	return
}

// file returns an attached log file for writing.
func (r *LogCollector) file(task *Task) (f *os.File, err error) {
	f, found, err := r.find(task)
	if found || err != nil {
		return
	}
	f, err = r.create(task)
	return
}

// find finds and opens an attached log file by name.
func (r *LogCollector) find(task *Task) (f *os.File, found bool, err error) {
	var file model.File
	name := r.name()
	for _, attached := range task.Attached {
		if attached.Name == name {
			found = true
			err = r.DB.First(&file, attached.ID).Error
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
		}
	}
	if !found {
		return
	}
	f, err = os.OpenFile(file.Path, os.O_RDONLY|os.O_APPEND, 0666)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	st, err := f.Stat()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	r.nSkip = st.Size()
	return
}

// create creates and attaches the log file.
func (r *LogCollector) create(task *Task) (f *os.File, err error) {
	file := &model.File{Name: r.name()}
	err = r.DB.Create(file).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	f, err = os.Create(file.Path)
	if err != nil {
		_ = r.DB.Delete(file)
		err = liberr.Wrap(err)
		return
	}
	task.attach(file)
	return
}

// copy data.
// The read bytes are discarded when smaller than nSkip.
// The offset is adjusted when to account for the buffer
// containing bytes to be skipped and written.
func (r *LogCollector) copy(reader io.ReadCloser, writer io.Writer, ctx context.Context) (err error) {
	timer := time.NewTimer(time.Second)
	canceled := false
	readCh := make(chan []byte)
	if r.nBuf < 1 {
		r.nBuf = 4096
	}
	go func() {
		defer func() {
			close(readCh)
		}()
		for {
			buf := make([]byte, r.nBuf)
			n, err := reader.Read(buf)
			if err != nil {
				if canceled {
					return
				}
				if err != io.EOF {
					Log.Error(err, "")
				}
				break
			}
			readCh <- buf[:n]
		}
	}()
	for {
		var buf []byte
		nRead := int64(-1)
		select {
		case <-ctx.Done():
			canceled = true
			_ = reader.Close()
			return
		case b, read := <-readCh:
			if read {
				nRead = int64(len(b))
				buf = b
			}
		case <-timer.C:
			nRead = 0
		}
		if nRead == -1 { // EOF.
			break
		}
		if nRead == 0 { // Timeout.
			continue
		}
		offset := int64(0)
		if r.nSkip > 0 {
			if nRead > r.nSkip {
				offset = r.nSkip
				r.nSkip = 0
			} else {
				r.nSkip -= nRead
				continue
			}
		}
		b := buf[offset:]
		_, err = writer.Write(b)
		if err != nil {
			return
		}
		if f, cast := writer.(*os.File); cast {
			err = f.Sync()
			if err != nil {
				return
			}
		}
	}
	return
}

// LogManager manages log collectors.
type LogManager struct { // LogCenter
	// collector registry.
	collector map[string]*LogCollector
	// mutex
	mutex sync.Mutex
	// DB
	DB *gorm.DB
}

// EnsureCollection - ensure each container has a log collector.
func (m *LogManager) EnsureCollection(task *Task, pod *core.Pod, ctx context.Context) (err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for _, container := range pod.Status.ContainerStatuses {
		if container.State.Waiting != nil {
			continue
		}
		collector := &LogCollector{
			Owner:     m,
			Registry:  m.collector,
			DB:        m.DB,
			Pod:       pod,
			Container: &container,
		}
		key := collector.key()
		if _, found := m.collector[key]; found {
			continue
		}
		err = collector.Begin(task, ctx)
		if err != nil {
			return
		}
		m.collector[key] = collector
	}
	return
}

// terminated provides notification that a collector has terminated.
func (m *LogManager) terminated(collector *LogCollector) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for i := range m.collector {
		if collector == m.collector[i] {
			key := collector.key()
			delete(m.collector, key)
			break
		}
	}
}
