package settings

import (
	"os"
	"strconv"
)

const (
	EnvNamespace         = "NAMESPACE"
	EnvDbPath            = "DB_PATH"
	EnvDbSeedPath        = "DB_SEED_PATH"
	EnvBucketPath        = "BUCKET_PATH"
	EnvRwxSupported      = "RWX_SUPPORTED"
	EnvCachePath         = "CACHE_PATH"
	EnvCachePvc          = "CACHE_PVC"
	EnvPassphrase        = "ENCRYPTION_PASSPHRASE"
	EnvTaskReapCreated   = "TASK_REAP_CREATED"
	EnvTaskReapSucceeded = "TASK_REAP_SUCCEEDED"
	EnvTaskReapFailed    = "TASK_REAP_FAILED"
	EnvTaskSA            = "TASK_SA"
	EnvTaskRetries       = "TASK_RETRIES"
	EnvFrequencyTask     = "FREQUENCY_TASK"
	EnvFrequencyReaper   = "FREQUENCY_REAPER"
	EnvDevelopment       = "DEVELOPMENT"
	EnvFileTTL           = "FILE_TTL"
)

type Hub struct {
	// k8s namespace.
	Namespace string
	// DB settings.
	DB struct {
		Path     string
		SeedPath string
	}
	// Bucket settings.
	Bucket struct {
		Path string
	}
	// File settings.
	File struct {
		TTL int
	}
	// Cache settings.
	Cache struct {
		RWX  bool
		Path string
		PVC  string
	}
	// Encryption settings.
	Encryption struct {
		Passphrase string
	}
	// Task
	Task struct {
		SA      string
		Retries int
		Reaper  struct { // minutes.
			Created   int
			Succeeded int
			Failed    int
		}
	}
	// Frequency
	Frequency struct {
		Task   int
		Reaper int
		Volume int
	}
	// Development environment
	Development bool
}

func (r *Hub) Load() (err error) {
	var found bool
	r.Namespace, err = r.namespace()
	if err != nil {
		return
	}
	r.DB.Path, found = os.LookupEnv(EnvDbPath)
	if !found {
		r.DB.Path = "/tmp/tackle.db"
	}
	r.DB.SeedPath, found = os.LookupEnv(EnvDbSeedPath)
	if !found {
		r.DB.SeedPath = "/tmp/seed"
	}
	r.Bucket.Path, found = os.LookupEnv(EnvBucketPath)
	if !found {
		r.Bucket.Path = "/tmp/bucket"
	}
	s, found := os.LookupEnv(EnvRwxSupported)
	if found {
		b, _ := strconv.ParseBool(s)
		r.Cache.RWX = b
	} else {
		r.Cache.RWX = true
	}
	r.Cache.PVC, found = os.LookupEnv(EnvCachePvc)
	if !found {
		r.Cache.PVC = "cache"
	}
	r.Cache.Path, found = os.LookupEnv(EnvCachePath)
	if !found {
		r.Cache.Path = "/cache"
	}
	r.Encryption.Passphrase, found = os.LookupEnv(EnvPassphrase)
	if !found {
		r.Encryption.Passphrase = "tackle"
	}
	s, found = os.LookupEnv(EnvTaskReapCreated)
	if found {
		n, _ := strconv.Atoi(s)
		r.Task.Reaper.Created = n
	} else {
		r.Task.Reaper.Created = 4320 // 72 hours.
	}
	s, found = os.LookupEnv(EnvTaskReapSucceeded)
	if found {
		n, _ := strconv.Atoi(s)
		r.Task.Reaper.Succeeded = n
	} else {
		r.Task.Reaper.Succeeded = 60 // 1 hour.
	}
	s, found = os.LookupEnv(EnvTaskReapFailed)
	if found {
		n, _ := strconv.Atoi(s)
		r.Task.Reaper.Failed = n
	} else {
		r.Task.Reaper.Failed = 4320 // 72 hours.
	}
	r.Task.SA, found = os.LookupEnv(EnvTaskSA)
	if !found {
		r.Task.SA = "tackle-hub"
	}
	s, found = os.LookupEnv(EnvTaskRetries)
	if found {
		n, _ := strconv.Atoi(s)
		r.Task.Retries = n
	} else {
		r.Task.Retries = 1
	}
	s, found = os.LookupEnv(EnvFrequencyTask)
	if found {
		n, _ := strconv.Atoi(s)
		r.Frequency.Task = n
	} else {
		r.Frequency.Task = 1 // 1 second.
	}
	s, found = os.LookupEnv(EnvFrequencyReaper)
	if found {
		n, _ := strconv.Atoi(s)
		r.Frequency.Reaper = n
	} else {
		r.Frequency.Reaper = 1 // 1 minute.
	}
	s, found = os.LookupEnv(EnvDevelopment)
	if found {
		b, _ := strconv.ParseBool(s)
		r.Development = b
	} else {
		r.Development = false
	}
	s, found = os.LookupEnv(EnvFileTTL)
	if found {
		n, _ := strconv.Atoi(s)
		r.File.TTL = n
	} else {
		r.File.TTL = 720 // 12 hours.
	}

	return
}

//
// namespace determines the namespace.
func (r *Hub) namespace() (ns string, err error) {
	ns, found := os.LookupEnv(EnvNamespace)
	if found {
		return
	}
	path := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	b, err := os.ReadFile(path)
	if err == nil {
		ns = string(b)
		return
	}
	if os.IsNotExist(err) {
		ns = "konveyor-tackle"
		err = nil
	}

	return
}
