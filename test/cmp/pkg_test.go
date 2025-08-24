package cmp

import (
	"testing"

	"github.com/konveyor/tackle2-hub/migration/json"
)

func TestCmp(t *testing.T) {
	eq, d := Eq("hello", "world")
	print(eq)
	print("\n")
	print(d)

	a := []int{1, 2, 3}
	b := []int{1, 3, 2}
	eq, d = Eq(a, b)
	print(eq)
	print("\n")
	print(d)

	print("\n")

	a = []int{1, 2, 3, 4}
	b = []int{1, 3, 2}
	eq, d = Eq(a, b)
	print(eq)
	print("\n")
	print(d)
	print("\n")

	a = []int{1, 3, 2}
	b = []int{1, 2, 3, 4, 5}
	eq, d = Eq(a, b)
	print(eq)
	print("\n")
	print(d)
	print("\n")

	tA := T{ID: 1, Thing: &T2{Name: "Roger", Age: 18, List: b}, List: a}
	tB := T{ID: 1, Thing: &T2{Name: "Brian", Age: 18}, List: b}
	eq, d = Eq(tA, tB)
	print(eq)
	print("\n")
	print(d)
	print("\n")

	eq, d = Eq(nil, 10)
	print(eq)
	print("\n")
	print(d)
	print("\n")

	mA := json.Map{
		"id":   1,
		"name": "Roger",
		"age":  18,
	}
	mB := json.Map{
		"id":   1,
		"name": "Brian",
		"age":  18,
	}
	eq, d = Eq(mA, mB)
	print(eq)
	print("\n")
	print(d)
	print("\n")
}

type T struct {
	ID    int
	List  []int
	Thing *T2
}

type T2 struct {
	Name string
	Age  int
	List []int
}
