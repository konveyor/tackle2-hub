package secret

import (
	"errors"
	"fmt"
	"reflect"
)

// Secret provides encryption of objects.
type Secret struct {
	Cipher Cipher
}

// Encrypt object.
// When object is:
// - *string - the string is encrypted.
// - struct - (string) fields with `secret:` tag are encrypted.
// - map[string]any - string fields are encrypted.
func (r Secret) Encrypt(object any) (err error) {
	err = r.Update(object, func(in string) (out string) {
		if r.isEncrypted(r.Cipher, in) {
			out = in
		} else {
			out, err = r.Cipher.Encrypt(in)
			if err != nil {
				panic(err)
			}
		}
		return
	})
	return
}

// Decrypt object.
// When object is:
// - *string - the string is decrypted.
// - struct - (string) fields with `secret:` tag are decrypted.
// - map[string]any - string fields are decrypted.
func (r Secret) Decrypt(object any) (err error) {
	err = r.Update(object, func(in string) (out string) {
		out, err = r.Cipher.Decrypt(in)
		if err != nil {
			panic(err)
		}
		return
	})
	return
}

// Update updates the specifeed
func (r Secret) Update(object any, fn func(string) string) (err error) {
	defer func() {
		p := recover()
		if p != nil {
			switch p.(type) {
			case error:
				err = p.(error)
			default:
				err = errors.New(fmt.Sprint(p))
			}
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

func (r Secret) isEncrypted(cipher Cipher, v string) (encrypted bool) {
	_, err := cipher.Decrypt(v)
	encrypted = err == nil
	return
}
