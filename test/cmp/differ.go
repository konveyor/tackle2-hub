package cmp

import (
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

type Differ struct {
	IgnoredPaths []string
	//
	path  []string
	kinds []reflect.Kind
	notes []string
}

func (d *Differ) Eq(expected, got any) (eq bool, report string) {
	eq, report = d.Inspect(expected, got)
	if eq {
		return
	}
	sep := "\n__________________________\n"
	report = fmt.Sprintf(
		"\nExpected:%s%s\n\nGot:%s%s\nDiff%s%s\n\n",
		sep,
		Format(expected),
		sep,
		Format(got),
		sep,
		report)
	return
}

func (d *Differ) Inspect(a, b any) (eq bool, report string) {
	d.reset()
	d.cmp(a, b)
	report = strings.Join(d.notes, "\n")
	eq = len(d.notes) == 0
	return
}

func (d *Differ) reset() {
	d.path = nil
	d.notes = nil
	d.kinds = nil
}

func (d *Differ) push(k reflect.Kind, p string, v ...any) {
	d.kinds = append(d.kinds, k)
	p = fmt.Sprintf(p, v...)
	if len(d.path) == 0 {
		p = strings.TrimPrefix(p, ".")
	}
	d.path = append(d.path, p)
}

func (d *Differ) pop() {
	if len(d.path) > 0 {
		d.path = d.path[:len(d.path)-1]
	}
	if len(d.kinds) > 0 {
		d.kinds = d.kinds[:len(d.kinds)-1]
	}
}

func (d *Differ) note(n string, v ...any) {
	parts := make([]string, 0, len(d.path))
	for _, p := range d.path {
		if strings.HasPrefix(p, "[") {
			continue
		}
		parts = append(parts, p)
	}
	current := strings.Join(parts, "")
	for _, p := range d.IgnoredPaths {
		matched, _ := filepath.Match(p, current)
		if matched {
			return
		}
	}
	d.notes = append(
		d.notes,
		fmt.Sprintf(n, v...))
}

func (d *Differ) cmpNIL(a, b any) (n bool) {
	var nA, nB bool
	switch v := a.(type) {
	case reflect.Value:
		nA = !v.IsValid()
	default:
		nA = a == nil
	}
	switch v := b.(type) {
	case reflect.Value:
		nB = !v.IsValid()
	default:
		nB = b == nil
	}
	if nA || nB {
		if nA && !nB {
			d.note(
				"~ %s%s<nil> expected: <ptr>",
				d.operator(),
				d.at())
		}
		if nB && !nA {
			d.note(
				"~ %s%s<ptr> expected: <nil>",
				d.operator(),
				d.at())
		}
	}
	n = nA || nB
	return
}

func (d *Differ) at() (path string) {
	path = strings.Join(d.path, "")
	return
}

func (d *Differ) kind() (k reflect.Kind) {
	n := len(d.kinds)
	if n == 0 {
		return
	}
	k = d.kinds[n-1]
	return
}

func (d *Differ) operator() (op string) {
	switch d.kind() {
	case reflect.Map:
		op = ": "
	default:
		op = " = "
	}
	return
}

func (d *Differ) cmp(a, b any) {
	if d.cmpNIL(a, b) {
		return
	}
	tA := reflect.TypeOf(a)
	tB := reflect.TypeOf(b)
	vA := reflect.ValueOf(a)
	vB := reflect.ValueOf(b)
	if tA.Kind() == reflect.Ptr {
		tA = tA.Elem()
		vA = vA.Elem()
	}
	if tB.Kind() == reflect.Ptr {
		tB = tB.Elem()
		vB = vB.Elem()
	}
	if tA != tB {
		d.note(
			"%s: (type) %s != %s",
			d.at(),
			tA.Name(),
			tB.Name())
		return
	}
	kind := tA.Kind()
	switch kind {
	case reflect.Slice:
		for i := 0; i < vA.Len(); i++ {
			d.push(kind, "[%d]", i)
			xA := vA.Index(i).Interface()
			if i < vB.Len() {
				xB := vB.Index(i).Interface()
				d.cmp(xA, xB)
			} else {
				d.note(
					"- %s = %#v",
					d.at(),
					xA)
			}
			d.pop()
		}
		for i := vA.Len(); i < vB.Len(); i++ {
			d.push(kind, "[%d]", i)
			xB := vB.Index(i).Interface()
			d.note(
				"+ %s = %#v",
				d.at(),
				xB)
			d.pop()
		}
	case reflect.Map:
		keyset := vA.MapKeys()
		sort.Slice(
			keyset, func(i, j int) bool {
				return keyset[i].String() < keyset[j].String()
			})
		for _, kA := range keyset {
			d.push(kind, ".%s", kA.String())
			vA := vA.MapIndex(kA)
			vB := vB.MapIndex(kA)
			if d.cmpNIL(vA, vB) {
				continue
			}
			xA := vA.Interface()
			if !vB.IsValid() {
				d.note(
					"- %s: <value>",
					d.at())
			} else {
				xB := vB.Interface()
				d.cmp(xA, xB)
			}
			d.pop()
		}
		keyset = vB.MapKeys()
		sort.Slice(
			keyset, func(i, j int) bool {
				return keyset[i].String() < keyset[j].String()
			})
		for _, kB := range keyset {
			d.push(kind, ".%s", kB.String())
			vA := vA.MapIndex(kB)
			vB := vB.MapIndex(kB)
			if d.cmpNIL(vA, vB) {
				continue
			}
			if !vA.IsValid() {
				d.note(
					"+ %s: <value>",
					d.at())
			}
			d.pop()
		}
	case reflect.Struct:
		for i := 0; i < vA.NumField(); i++ {
			ftA := tA.Field(i)
			if !ftA.IsExported() {
				continue
			}
			fA := vA.Field(i)
			fB := vB.Field(i)
			if fA.Kind() == reflect.Ptr {
				fA = fA.Elem()
			}
			if fB.Kind() == reflect.Ptr {
				fB = fB.Elem()
			}
			name := tA.Field(i).Name
			if !ftA.Anonymous {
				d.push(kind, ".%s", name)
			}
			if !d.cmpNIL(fA, fB) {
				xA := fA.Interface()
				xB := fB.Interface()
				d.cmp(xA, xB)
			}
			if !ftA.Anonymous {
				d.pop()
			}
		}
	default:
		if !reflect.DeepEqual(a, b) {
			xA := vA.Interface()
			xB := vB.Interface()
			d.note(
				"~ %s%s%#v expected: %#v",
				d.at(),
				d.operator(),
				xB,
				xA)
		}
	}
}
