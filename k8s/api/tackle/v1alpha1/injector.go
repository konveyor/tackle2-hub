package v1alpha1

type Field struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Key  string `json:"key"`
}

type Injector struct {
	Kind   string  `json:"kind"`
	Fields []Field `json:"fields"`
}
