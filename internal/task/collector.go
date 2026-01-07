package task

import (
	"context"
	"io"
	"os"
	"sync"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/k8s"
	"github.com/konveyor/tackle2-hub/internal/model"
	"gorm.io/gorm"
	core "k8s.io/api/core/v1"
)

// LogManager manages log collectors.
type LogManager struct {
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
	for i := range pod.Status.ContainerStatuses {
		container := &pod.Status.ContainerStatuses[i]
		if container.State.Waiting != nil {
			continue
		}
		collector := &LogCollector{
			Owner:     m,
			DB:        m.DB,
			Pod:       pod,
			Container: container,
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
	delete(m.collector, collector.key())
}

// LogCollector collect and report container logs.
type LogCollector struct {
	Owner     *LogManager
	DB        *gorm.DB
	Pod       *core.Pod
	Container *core.ContainerStatus
	//
	nBuf  int
	nSkip int64
}

// Begin - get container log and store in file.
// - Create file resource and attach to the task.
// - Request logs.
// - Write (copy) log.
// - Unregister self.
func (r *LogCollector) Begin(task *Task, ctx context.Context) (err error) {
	f, err := r.file(task)
	if err != nil {
		return
	}
	go func() {
		defer func() {
			r.Owner.terminated(r)
			_ = f.Close()
		}()
		reader, err := r.request(ctx)
		if err != nil {
			return
		}
		defer func() {
			_ = reader.Close()
		}()
		err = r.copy(reader, f)
		Log.Error(err, "")
	}()
	return
}

// key returns the collector key.
func (r *LogCollector) key() (key string) {
	key = r.Pod.Name + "." + r.Container.Name
	return
}

// request logs from k8s.
func (r *LogCollector) request(ctx context.Context) (reader io.ReadCloser, err error) {
	options := &core.PodLogOptions{
		Container: r.Container.Name,
		Follow:    true,
	}
	clientSet, err := k8s.NewClientSet()
	if err != nil {
		return
	}
	podClient := clientSet.CoreV1().Pods(Settings.Hub.Namespace)
	req := podClient.GetLogs(r.Pod.Name, options)
	reader, err = req.Stream(ctx)
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
func (r *LogCollector) copy(reader io.ReadCloser, writer io.Writer) (err error) {
	if r.nBuf < 1 {
		r.nBuf = 4096
	}
	buf := make([]byte, r.nBuf)
	for {
		n := 0
		n, err = reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		nRead := int64(n)
		if nRead == 0 {
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
		b := buf[offset:nRead]
		_, err = writer.Write(b)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		if f, cast := writer.(*os.File); cast {
			err = f.Sync()
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
		}
	}
	return
}
