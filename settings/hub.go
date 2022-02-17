package settings

import (
	"os"
)

const (
	EnvNamespace  = "NAMESPACE"
	EnvDbPath     = "DB_PATH"
	EnvDbSeedPath = "DB_SEED_PATH"
	EnvBucketPath = "BUCKET_PATH"
	EnvBucketPVC  = "BUCKET_PVC"
	EnvPassphrase = "ENCRYPTION_PASSPHRASE"
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
		PVC  string
	}
	// Encryption settings.
	Encryption struct {
		Passphrase string
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
	r.Bucket.PVC, found = os.LookupEnv(EnvBucketPVC)
	if !found {
		r.Bucket.PVC = "bucket"
	}
	r.Encryption.Passphrase, found = os.LookupEnv(EnvPassphrase)
	if !found {
		r.Encryption.Passphrase = "tackle"
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
		ns = "tackle-operator"
		err = nil
	}

	return
}
