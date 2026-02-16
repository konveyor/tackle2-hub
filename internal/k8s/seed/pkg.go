package seed

import (
	"embed"
	"reflect"
	"strings"

	crd "github.com/konveyor/tackle2-hub/internal/k8s/api/tackle/v1alpha1"
	core "k8s.io/api/core/v1"
	k8srt "k8s.io/apimachinery/pkg/runtime"
	k8scheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

//go:embed resources/*.yaml
var embedded embed.FS

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
		"resources/configmap.yaml": &core.ConfigMap{},
		"resources/tackle.yaml":    &crd.Tackle{},
		"resources/addon.yaml":     &crd.Addon{},
		"resources/extension.yaml": &crd.Extension{},
		"resources/task.yaml":      &crd.Task{},
		"resources/jsd.yaml":       &crd.Schema{},
	}
	for path, r := range files {
		b, err := embedded.ReadFile(path)
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
