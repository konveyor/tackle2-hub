package secret

import (
	"errors"
	"fmt"
	"reflect"
)

const (
	TagEncrypted = "encrypted"
	TagHashed    = "hashed"
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
	selector := func(f Field) func(string) string {
		switch f.tag {
		case "", TagEncrypted:
			return func(in string) (out string) {
				if r.isEncrypted(r.Cipher, in) {
					out = in
				} else {
					out, err = r.Cipher.Encrypt(in)
					if err != nil {
						panic(err)
					}
				}
				return
			}
		case TagHashed:
			return func(in string) (out string) {
				out = in
				return
			}
		default:
			panic("unknown tag: " + f.tag)
		}
	}
	_, err = r.Update(object, selector, Field{})
	return
}

// Decrypt object.
// When object is:
// - *string - the string is decrypted.
// - struct - (string) fields with `secret:` tag are decrypted.
// - map[string]any - string fields are decrypted.
func (r Secret) Decrypt(object any) (err error) {
	selector := func(f Field) func(string) string {
		switch f.tag {
		case "", TagEncrypted:
			return func(in string) (out string) {
				out, err = r.Cipher.Decrypt(in)
				if err != nil {
					panic(err)
				}
				return
			}
		case TagHashed:
			return func(in string) (out string) {
				out = in
				return
			}
		default:
			panic("unknown tag: " + f.tag)
		}
	}
	_, err = r.Update(object, selector, Field{})
	return
}

// Encode object.
// When object is:
// - *string - the string is encrypted.
// - struct - (string) fields with `secret:` tag are encoded based on tag (value).
// - map[string]any - string fields are encrypted.
func (r Secret) Encode(object any) (fields []Field, err error) {
	selector := func(f Field) func(string) string {
		switch f.tag {
		case "", TagEncrypted:
			return func(in string) (out string) {
				if r.isEncrypted(r.Cipher, in) {
					out = in
				} else {
					out, err = r.Cipher.Encrypt(in)
					if err != nil {
						panic(err)
					}
				}
				return
			}
		case TagHashed:
			return func(in string) (out string) {
				if isHashedPassword(in) {
					out = in
				} else {
					out = HashPassword(in)
				}
				return
			}
		default:
			panic("unknown tag: " + f.tag)
		}
	}
	fields, err = r.Update(object, selector, Field{})
	return
}

// Decode object.
// When object is:
// - *string - the string is decrypted.
// - struct - (string) fields with `secret:` tag are decoded based on tag (value).
// - map[string]any - string fields are decrypted.
func (r Secret) Decode(object any) (err error) {
	selector := func(f Field) func(string) string {
		switch f.tag {
		case "", TagEncrypted:
			return func(in string) (out string) {
				out, err = r.Cipher.Decrypt(in)
				if err != nil {
					panic(err)
				}
				return
			}
		case TagHashed:
			return func(in string) (out string) {
				out = in
				return
			}
		default:
			panic("unknown tag: " + f.tag)
		}
	}
	_, err = r.Update(object, selector, Field{})
	return
}

// Update updates the specified object.
func (r Secret) Update(object any, selector Selector, field Field) (fields []Field, err error) {
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
		fn := selector(field)
		mv.SetString(fn(mv.String()))
	case reflect.Struct:
		for i := 0; i < mt.NumField(); i++ {
			ft := mt.Field(i)
			fv := mv.Field(i)
			sf := r.field(&field, ft, fv)
			if !sf.hasTag || !sf.exported {
				continue
			}
			switch fv.Kind() {
			case reflect.String:
				fields = append(fields, sf)
				fn := selector(sf)
				s := fn(fv.String())
				fv.SetString(s)
			case reflect.Map,
				reflect.Struct,
				reflect.Slice:
				var nested []Field
				nested, err = r.Update(fv.Interface(), selector, sf)
				fields = append(fields, nested...)
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
				fn := selector(field)
				s := fn(v.String())
				mv.SetMapIndex(k, reflect.ValueOf(s))
			case reflect.Map,
				reflect.Struct,
				reflect.Slice:
				var nested []Field
				nested, err = r.Update(v.Interface(), selector, field)
				fields = append(fields, nested...)
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
				fn := selector(field)
				s := fn(v.String())
				v.SetString(s)
			case reflect.Map,
				reflect.Struct,
				reflect.Slice:
				var nested []Field
				nested, err = r.Update(v.Interface(), selector, field)
				fields = append(fields, nested...)
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

// Selector defines encoding/decoding methods.
type Selector func(Field) func(string) string

// Field represents a visited struct field.
type Field struct {
	parent   *Field
	hasTag   bool
	exported bool
	name     string
	tag      string
	secret   string
}

// Fqn returns the fully qualified field name.
func (f Field) Fqn() (n string) {
	n = f.name
	if f.parent != nil && f.parent.name != "" {
		n = f.parent.Fqn() + "." + n
	}
	return
}

// Root returns true when the field is a root field.
func (f Field) Root() (b bool) {
	b = f.parent == nil || f.parent.name == ""
	return
}

// Secret returns the field secret value.
func (f Field) Secret() (s string) {
	s = f.secret
	return
}

func (r Secret) field(parent *Field, ft reflect.StructField, fv reflect.Value) (f Field) {
	f.parent = parent
	f.name = ft.Name
	f.exported = ft.IsExported()
	if fv.Kind() == reflect.String {
		f.secret = fv.String()
	}
	v, found := ft.Tag.Lookup("secret")
	if found {
		f.hasTag = true
		f.tag = v
	}
	return
}

func (r Secret) isEncrypted(cipher Cipher, v string) (encrypted bool) {
	_, err := cipher.Decrypt(v)
	encrypted = err == nil
	return
}
