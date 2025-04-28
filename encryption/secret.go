package encryption

import (
	"reflect"
)

type Secret struct {
	Passphrase string
}

func (r Secret) Encrypt(object any) (err error) {
	en := &AESGCM{}
	en.Use(r.Passphrase)
	r.Find(object, func(in string) (out string, err error) {
		if r.isEncrypted(en, in) {
			out = in
		} else {
			out, err = en.Encrypt(in)
		}
		return
	})
	return
}

func (r Secret) Decrypt(object any) (err error) {
	en := &AESGCM{}
	en.Use(r.Passphrase)
	r.Find(object, en.Decrypt)
	return
}

func (r Secret) Find(object any, fn func(string) (string, error)) {
	mt := reflect.TypeOf(object)
	mv := reflect.ValueOf(object)
	if mt.Kind() == reflect.Ptr {
		mt = mt.Elem()
		mv = mv.Elem()
	}
	switch mv.Kind() {
	case reflect.Struct:
		for i := 0; i < mt.NumField(); i++ {
			ft := mt.Field(i)
			fv := mv.Field(i)
			if !r.hasTag(ft) || !ft.IsExported() {
				continue
			}
			switch fv.Kind() {
			case reflect.String:
				s, err := fn(fv.String())
				if err != nil {
					return
				}
				fv.SetString(s)
			case reflect.Map,
				reflect.Struct,
				reflect.Slice:
				r.Find(fv.Interface(), fn)
			default:
			}
		}
	case reflect.Map:
		for _, k := range mv.MapKeys() {
			v := mv.MapIndex(k)
			v = v.Elem()
			switch v.Kind() {
			case reflect.String:
				s, err := fn(v.String())
				if err != nil {
					return
				}
				mv.SetMapIndex(k, reflect.ValueOf(s))
			case reflect.Map,
				reflect.Struct,
				reflect.Slice:
				r.Find(v.Interface(), fn)
			default:
			}
		}
	case reflect.Slice:
		for i := 0; i < mv.Len(); i++ {
			v := mv.Index(i)
			switch v.Kind() {
			case reflect.String:
				s, err := fn(v.String())
				if err != nil {
					return
				}
				v.SetString(s)
			case reflect.Map,
				reflect.Struct,
				reflect.Slice:
				r.Find(v.Interface(), fn)
			default:
			}
		}
	default:
	}
}

func (r Secret) hasTag(fv reflect.StructField) (found bool) {
	_, found = fv.Tag.Lookup("secret")
	return
}

func (r Secret) isEncrypted(en *AESGCM, v string) (encrypted bool) {
	_, err := en.Decrypt(v)
	encrypted = err == nil
	return
}
