package settings

import (
	"os"
	"strconv"
	"time"
)

const (
	EnvNamespace               = "NAMESPACE"
	EnvDbPath                  = "DB_PATH"
	EnvDbSeedPath              = "DB_SEED_PATH"
	EnvBucketPath              = "BUCKET_PATH"
	EnvRwxSupported            = "RWX_SUPPORTED"
	EnvCachePath               = "CACHE_PATH"
	EnvCachePvc                = "CACHE_PVC"
	EnvSharedPath              = "SHARED_PATH"
	EnvPassphrase              = "ENCRYPTION_PASSPHRASE"
	EnvTaskReapCreated         = "TASK_REAP_CREATED"
	EnvTaskReapSucceeded       = "TASK_REAP_SUCCEEDED"
	EnvTaskReapFailed          = "TASK_REAP_FAILED"
	EnvTaskSA                  = "TASK_SA"
	EnvTaskRetries             = "TASK_RETRIES"
	EnvTaskPreemptEnabled      = "TASK_PREEMPT_ENABLED"
	EnvTaskPreemptDelayed      = "TASK_PREEMPT_DELAYED"
	EnvTaskPreemptPostponed    = "TASK_PREEMPT_POSTPONED"
	EnvTaskPreemptRate         = "TASK_PREEMPT_RATE"
	EnvFrequencyTask           = "FREQUENCY_TASK"
	EnvFrequencyReaper         = "FREQUENCY_REAPER"
	EnvDevelopment             = "DEVELOPMENT"
	EnvBucketTTL               = "BUCKET_TTL"
	EnvFileTTL                 = "FILE_TTL"
	EnvAppName                 = "APP_NAME"
	EnvDisconnected            = "DISCONNECTED"
	EnvAnalysisReportPath      = "ANALYSIS_REPORT_PATH"
	EnvAnalysisArchiverEnabled = "ANALYSIS_ARCHIVER_ENABLED"
	EnvDiscoveryEnabled        = "DISCOVERY_ENABLED"
	EnvDiscoveryLabel          = "DISCOVERY_LABEL"
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
		TTL  int
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
	// Shared mount settings.
	Shared struct {
		Path string
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
		Preemption struct { // seconds.
			Enabled   bool
			Delayed   time.Duration
			Postponed time.Duration
			Rate      int
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
	// Product - deployed as product.
	Product bool
	// Disconnected indicates no cluster.
	Disconnected bool
	// Analysis settings.
	Analysis struct {
		ReportPath      string
		ArchiverEnabled bool
	}
	// Discovery settings.
	Discovery struct {
		Enabled bool
		Label   string
	}
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
	}
	r.Cache.PVC, found = os.LookupEnv(EnvCachePvc)
	if !found {
		r.Cache.PVC = "cache"
	}
	r.Cache.Path, found = os.LookupEnv(EnvCachePath)
	if !found {
		r.Cache.Path = "/cache"
	}
	r.Shared.Path, found = os.LookupEnv(EnvSharedPath)
	if !found {
		r.Shared.Path = "/shared"
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
	s, found = os.LookupEnv(EnvTaskPreemptEnabled)
	if found {
		b, _ := strconv.ParseBool(s)
		r.Task.Preemption.Enabled = b
	}
	s, found = os.LookupEnv(EnvTaskPreemptDelayed)
	if found {
		n, _ := strconv.Atoi(s)
		r.Task.Preemption.Delayed = time.Duration(n) * time.Second
	} else {
		r.Task.Preemption.Delayed = time.Minute
	}
	s, found = os.LookupEnv(EnvTaskPreemptPostponed)
	if found {
		n, _ := strconv.Atoi(s)
		r.Task.Preemption.Postponed = time.Duration(n) * time.Second
	} else {
		r.Task.Preemption.Postponed = time.Minute
	}
	s, found = os.LookupEnv(EnvTaskPreemptRate)
	if found {
		n, _ := strconv.Atoi(s)
		if n < 0 {
			n = 0
		}
		if n > 100 {
			n = 100
		}
		r.Task.Preemption.Rate = n
	} else {
		r.Task.Preemption.Rate = 10
	}
	s, found = os.LookupEnv(EnvDevelopment)
	if found {
		b, _ := strconv.ParseBool(s)
		r.Development = b
	}
	s, found = os.LookupEnv(EnvBucketTTL)
	if found {
		n, _ := strconv.Atoi(s)
		r.Bucket.TTL = n
	} else {
		r.Bucket.TTL = 1 // minutes.
	}
	s, found = os.LookupEnv(EnvFileTTL)
	if found {
		n, _ := strconv.Atoi(s)
		r.File.TTL = n
	} else {
		r.File.TTL = 720 // minutes: 12 hours.
	}
	s, found = os.LookupEnv(EnvAppName)
	if found {
		r.Product = !(s == "" || s == "tackle")
	}
	s, found = os.LookupEnv(EnvDisconnected)
	if found {
		b, _ := strconv.ParseBool(s)
		r.Disconnected = b
	}
	r.Analysis.ReportPath, found = os.LookupEnv(EnvAnalysisReportPath)
	if !found {
		r.Analysis.ReportPath = "/tmp/analysis/report"
	}
	s, found = os.LookupEnv(EnvAnalysisArchiverEnabled)
	if found {
		b, _ := strconv.ParseBool(s)
		r.Analysis.ArchiverEnabled = b
	} else {
		r.Analysis.ArchiverEnabled = true
	}
	s, found = os.LookupEnv(EnvDiscoveryEnabled)
	if found {
		b, _ := strconv.ParseBool(s)
		r.Discovery.Enabled = !r.Disconnected && b
	} else {
		r.Discovery.Enabled = !r.Disconnected
	}
	s, found = os.LookupEnv(EnvDiscoveryLabel)
	if found {
		r.Discovery.Label = s
	} else {
		r.Discovery.Label = "konveyor.io/discovery"
	}

	return
}

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
