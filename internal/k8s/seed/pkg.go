package seed

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	crd "github.com/konveyor/tackle2-hub/internal/k8s/api/tackle/v1alpha1"
	k8srt "k8s.io/apimachinery/pkg/runtime"
	k8scheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// Scheme creates a scheme with both core Kubernetes types and custom CRDs.
func Scheme() (scheme *k8srt.Scheme) {
	scheme = k8srt.NewScheme()
	_ = k8scheme.AddToScheme(scheme)
	_ = crd.SchemeBuilder.AddToScheme(scheme)
	return
}

// Resources returns seed resources.
func Resources() []client.Object {
	var objects []client.Object
	files := map[string]client.Object{
		"tackle.yaml":    &crd.Tackle{},
		"addon.yaml":     &crd.Addon{},
		"extension.yaml": &crd.Extension{},
		"task.yaml":      &crd.Task{},
		"jsd.yaml":       &crd.Schema{},
	}
	for path, r := range files {
		path = filepath.Join(dataDir(), path)
		b, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		rt := reflect.TypeOf(r).Elem()
		content := strings.Split(string(b), "\n---\n")
		for _, d := range content {
			nt := reflect.New(rt)
			obj := nt.Interface().(client.Object)
			err = yaml.Unmarshal([]byte(d), obj)
			if err != nil {
				panic(err)
			}
			objects = append(objects, obj)
		}
	}
	return objects
}

// dataDir returns the path to the data directory.
func dataDir() (d string) {
	_, filename, _, _ := runtime.Caller(0)
	d = filepath.Join(filepath.Dir(filename), "resources")
	return
}
