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
	err = r.Update(object, func(in string) (out string) {
		if r.isEncrypted(en, in) {
			out = in
		} else {
			out, err = en.Encrypt(in)
			if err != nil {
				panic(err)
			}
		}
		return
	})
	return
}

func (r Secret) Decrypt(object any) (err error) {
	en := &AESGCM{}
	en.Use(r.Passphrase)
	err = r.Update(object, func(in string) (out string) {
		out, err = en.Decrypt(in)
		if err != nil {
			panic(err)
		}
		return
	})
	return
}

func (r Secret) Update(object any, fn func(string) string) (err error) {
	defer func() {
		p := recover()
		if p != nil {
			err = p.(error)
		}
	}()
	mt := reflect.TypeOf(object)
	mv := reflect.ValueOf(object)
	if mt.Kind() == reflect.Ptr {
		mt = mt.Elem()
		mv = mv.Elem()
	}
	switch mv.Kind() {
	case reflect.String:
		mv.SetString(fn(mv.String()))
	case reflect.Struct:
		for i := 0; i < mt.NumField(); i++ {
			ft := mt.Field(i)
			fv := mv.Field(i)
			if !r.hasTag(ft) || !ft.IsExported() {
				continue
			}
			switch fv.Kind() {
			case reflect.String:
				s := fn(fv.String())
				fv.SetString(s)
			case reflect.Map,
				reflect.Struct,
				reflect.Slice:
				err = r.Update(fv.Interface(), fn)
				if err != nil {
					return
				}
			default:
			}
		}
	case reflect.Map:
		for _, k := range mv.MapKeys() {
			v := mv.MapIndex(k)
			v = v.Elem()
			switch v.Kind() {
			case reflect.String:
				s := fn(v.String())
				mv.SetMapIndex(k, reflect.ValueOf(s))
			case reflect.Map,
				reflect.Struct,
				reflect.Slice:
				err = r.Update(v.Interface(), fn)
				if err != nil {
					return
				}
			default:
			}
		}
	case reflect.Slice:
		for i := 0; i < mv.Len(); i++ {
			v := mv.Index(i)
			switch v.Kind() {
			case reflect.String:
				s := fn(v.String())
				v.SetString(s)
			case reflect.Map,
				reflect.Struct,
				reflect.Slice:
				err = r.Update(v.Interface(), fn)
				if err != nil {
					return
				}
			default:
			}
		}
	default:
	}
	return
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
