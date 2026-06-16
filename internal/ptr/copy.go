package ptr

import "reflect"

// Copy returns a deep copy of pA.
func Copy[T any](pA *T) (pB *T) {
	if pA == nil {
		return
	}
	pB = new(T)
	vA := reflect.ValueOf(pA).Elem()
	vB := reflect.ValueOf(pB).Elem()
	copyValue(vA, vB)
	return
}

// copyValue returns a deep copied reflect value.
func copyValue(vA, vB reflect.Value) {
	switch vA.Kind() {
	case reflect.Ptr:
		if vA.IsNil() {
			vB.Set(reflect.Zero(vB.Type()))
			return
		}
		if vB.IsNil() {
			vB.Set(reflect.New(vA.Type().Elem()))
		}
		copyValue(vA.Elem(), vB.Elem())

	case reflect.Slice:
		if vA.IsNil() {
			vB.Set(reflect.Zero(vB.Type()))
			return
		}
		vB.Set(reflect.MakeSlice(vA.Type(), vA.Len(), vA.Cap()))
		for i := 0; i < vA.Len(); i++ {
			copyValue(vA.Index(i), vB.Index(i))
		}

	case reflect.Map:
		if vA.IsNil() {
			vB.Set(reflect.Zero(vB.Type()))
			return
		}
		vB.Set(reflect.MakeMap(vA.Type()))
		for _, k := range vA.MapKeys() {
			v := vA.MapIndex(k)
			newV := reflect.New(v.Type()).Elem()
			copyValue(v, newV)
			vB.SetMapIndex(k, newV)
		}

	case reflect.Struct:
		for i := 0; i < vA.NumField(); i++ {
			if vB.Field(i).CanSet() {
				copyValue(vA.Field(i), vB.Field(i))
			}
		}

	case reflect.Array:
		for i := 0; i < vA.Len(); i++ {
			copyValue(vA.Index(i), vB.Index(i))
		}

	default:
		if vB.CanSet() {
			vB.Set(vA)
		}
	}
}
