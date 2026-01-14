package settings

import (
	"bufio"
	"os"
	"os/user"
	"strconv"
)

const (
	EnvNamespace               = "NAMESPACE"
	EnvBuild                   = "BUILD"
	EnvDbPath                  = "DB_PATH"
	EnvDbMaxCon                = "DB_MAX_CONNECTION"
	EnvDbSeedPath              = "DB_SEED_PATH"
	EnvBucketPath              = "BUCKET_PATH"
	EnvRwxSupported            = "RWX_SUPPORTED"
	EnvCachePvc                = "CACHE_PVC"
	EnvPassphrase              = "ENCRYPTION_PASSPHRASE"
	EnvTaskReapCreated         = "TASK_REAP_CREATED"
	EnvTaskReapSucceeded       = "TASK_REAP_SUCCEEDED"
	EnvTaskReapFailed          = "TASK_REAP_FAILED"
	EnvTaskPodRetainSucceeded  = "TASK_POD_RETAIN_SUCCEEDED"
	EnvTaskPodRetainFailed     = "TASK_POD_RETAIN_FAILED"
	EnvTaskPodQuota            = "TASK_POD_QUOTA"
	EnvTaskSA                  = "TASK_SA"
	EnvTaskRetries             = "TASK_RETRIES"
	EnvTaskUid                 = "TASK_UID"
	EnvTaskEnabled             = "TASK_ENABLED"
	EnvFrequencyTask           = "FREQUENCY_TASK"
	EnvFrequencyReaper         = "FREQUENCY_REAPER"
	EnvFrequencyHeap           = "FREQUENCY_HEAP"
	EnvDevelopment             = "DEVELOPMENT"
	EnvBucketTTL               = "BUCKET_TTL"
	EnvFileTTL                 = "FILE_TTL"
	EnvAppName                 = "APP_NAME"
	EnvDisconnected            = "DISCONNECTED"
	EnvAnalysisReportPath      = "ANALYSIS_REPORT_PATH"
	EnvAnalysisArchiverEnabled = "ANALYSIS_ARCHIVER_ENABLED"
	EnvDiscoveryLabel          = "DISCOVERY_LABEL"
	EnvLogMaster               = "LOG_MASTER"
	EnvLogMigration            = "LOG_MIGRATION"
	EnvLogWeb                  = "LOG_WEB"
	EnvLogReaper               = "LOG_REAPER"
	EnvLogTask                 = "LOG_TASK"
	EnvLogAuth                 = "LOG_AUTH"
	EnvLogHeap                 = "LOG_HEAP"
	EnvLogCmd                  = "LOG_CMD"
	EnvLogSsh                  = "LOG_SSH"
	EnvLogScm                  = "LOG_SCM"
)

type Hub struct {
	// build version.
	Build string
	// k8s namespace.
	Namespace string
	// DB settings.
	DB struct {
		Path          string
		MaxConnection int
		SeedPath      string
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
		RWX bool
		PVC string
	}
	// Encryption settings.
	Encryption struct {
		Passphrase string
	}
	// Task
	Task struct {
		Enabled bool
		SA      string
		Retries int
		Reaper  struct { // minutes.
			Created   int
			Succeeded int
			Failed    int
		}
		Pod struct {
			Quota     int
			Retention struct {
				Succeeded int
				Failed    int
			}
		}
		UID int64
	}
	// Frequency
	Frequency struct {
		Task   int
		Reaper int
		Heap   int
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
		Label string
	}
	// Logging settings
	Log struct {
		Master    int
		Migration int
		Web       int
		Reaper    int
		Task      int
		Heap      int
		Auth      int
		Command   int
		SCM       int
		SSH       int
	}
}

func (r *Hub) Load() (err error) {
	var found bool
	r.Build, err = r.build()
	if err != nil {
		return
	}
	r.Namespace, err = r.namespace()
	if err != nil {
		return
	}
	r.DB.Path, found = os.LookupEnv(EnvDbPath)
	if !found {
		r.DB.Path = "/tmp/tackle.db"
	}
	s, found := os.LookupEnv(EnvDbMaxCon)
	if found {
		n, _ := strconv.Atoi(s)
		r.DB.MaxConnection = n
	} else {
		r.DB.MaxConnection = 1
	}
	r.DB.SeedPath, found = os.LookupEnv(EnvDbSeedPath)
	if !found {
		r.DB.SeedPath = "/tmp/seed"
	}
	r.Bucket.Path, found = os.LookupEnv(EnvBucketPath)
	if !found {
		r.Bucket.Path = "/tmp/bucket"
	}
	s, found = os.LookupEnv(EnvRwxSupported)
	if found {
		b, _ := strconv.ParseBool(s)
		r.Cache.RWX = b
	}
	r.Cache.PVC, found = os.LookupEnv(EnvCachePvc)
	if !found {
		r.Cache.PVC = "cache"
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
		r.Task.Reaper.Succeeded = 4320 // 72 hours.
	}
	s, found = os.LookupEnv(EnvTaskReapFailed)
	if found {
		n, _ := strconv.Atoi(s)
		r.Task.Reaper.Failed = n
	} else {
		r.Task.Reaper.Failed = 43200 // 720 hours (30 days).
	}
	s, found = os.LookupEnv(EnvTaskPodQuota)
	if found {
		n, _ := strconv.Atoi(s)
		r.Task.Pod.Quota = n
	} else {
		r.Task.Pod.Quota = 0 // 0=disabled
	}
	s, found = os.LookupEnv(EnvTaskPodRetainSucceeded)
	if found {
		n, _ := strconv.Atoi(s)
		r.Task.Pod.Retention.Succeeded = n
	} else {
		r.Task.Pod.Retention.Succeeded = 1
	}
	s, found = os.LookupEnv(EnvTaskPodRetainFailed)
	if found {
		n, _ := strconv.Atoi(s)
		r.Task.Pod.Retention.Failed = n
	} else {
		r.Task.Pod.Retention.Failed = 4320 // 72 hours.
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
	s, found = os.LookupEnv(EnvFrequencyHeap)
	if found {
		n, _ := strconv.Atoi(s)
		r.Frequency.Heap = n // seconds.
	} else {
		r.Frequency.Heap = 240 // 4 minutes.
	}
	s, found = os.LookupEnv(EnvTaskUid)
	if found {
		var uid int64
		uid, err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			return
		}
		r.Task.UID = uid
	} else {
		var uid int64
		var hubUser *user.User
		hubUser, err = user.Current()
		if err != nil {
			return
		}
		uid, err = strconv.ParseInt(hubUser.Uid, 10, 64)
		if err != nil {
			return
		}
		r.Task.UID = uid
	}
	s, found = os.LookupEnv(EnvDisconnected)
	if found {
		b, _ := strconv.ParseBool(s)
		r.Disconnected = b
	}
	s, found = os.LookupEnv(EnvTaskEnabled)
	if found {
		b, _ := strconv.ParseBool(s)
		r.Task.Enabled = !r.Disconnected && b
	} else {
		r.Task.Enabled = !r.Disconnected
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
	s, found = os.LookupEnv(EnvDiscoveryLabel)
	if found {
		r.Discovery.Label = s
	} else {
		r.Discovery.Label = "konveyor.io/discovery"
	}
	s, found = os.LookupEnv(EnvLogMaster)
	if found {
		n, _ := strconv.Atoi(s)
		r.Log.Master = n
	}
	s, found = os.LookupEnv(EnvLogMigration)
	if found {
		n, _ := strconv.Atoi(s)
		r.Log.Migration = n
	} else {
		r.Log.Migration = r.Log.Master
	}
	s, found = os.LookupEnv(EnvLogWeb)
	if found {
		n, _ := strconv.Atoi(s)
		r.Log.Web = n
	} else {
		r.Log.Web = r.Log.Master
	}
	s, found = os.LookupEnv(EnvLogReaper)
	if found {
		n, _ := strconv.Atoi(s)
		r.Log.Reaper = n
	} else {
		r.Log.Reaper = r.Log.Master
	}
	s, found = os.LookupEnv(EnvLogTask)
	if found {
		n, _ := strconv.Atoi(s)
		r.Log.Task = n
	} else {
		r.Log.Task = r.Log.Master
	}
	s, found = os.LookupEnv(EnvLogHeap)
	if found {
		n, _ := strconv.Atoi(s)
		r.Log.Heap = n
	} else {
		r.Log.Heap = r.Log.Master
	}
	s, found = os.LookupEnv(EnvLogAuth)
	if found {
		n, _ := strconv.Atoi(s)
		r.Log.Auth = n
	} else {
		r.Log.Auth = r.Log.Master
	}
	s, found = os.LookupEnv(EnvLogCmd)
	if found {
		n, _ := strconv.Atoi(s)
		r.Log.Command = n
	} else {
		r.Log.Command = r.Log.Master
	}
	s, found = os.LookupEnv(EnvLogSsh)
	if found {
		n, _ := strconv.Atoi(s)
		r.Log.SSH = n
	} else {
		r.Log.SSH = r.Log.Master
	}
	s, found = os.LookupEnv(EnvLogScm)
	if found {
		n, _ := strconv.Atoi(s)
		r.Log.SCM = n
	} else {
		r.Log.SCM = r.Log.Master
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

// build returns the hub build version.
// This is expected to be the output of `git describe`.
// Examples:
//
//	v0.6.0-ea89gcd
//	v0.6.0
func (r *Hub) build() (version string, err error) {
	version, found := os.LookupEnv(EnvBuild)
	if found {
		return
	}
	f, err := os.Open("/etc/hub-build")
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
			return
		}
	}
	defer func() {
		_ = f.Close()
	}()
	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		version = scanner.Text()
	}
	return
}
