package cmp

import (
	"testing"
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

	eq, d = Eq(nil, 10)
	print(eq)
	print("\n")
	print(d)
	print("\n")

	mA := map[string]any{
		"ID":   1,
		"List": a,
		"Name": map[string]any{
			"First":  "Elmer",
			"Middle": "James",
			"Last":   "Fudd",
		},
		"Address": map[string]any{
			"Street":  "123",
			"City":    "Huntsville",
			"State":   "AL",
			"Country": "US",
			"Zip":     "12345",
		},
	}

	mB := map[string]any{
		"ID":   1,
		"List": b,
		"Name": map[string]any{
			"First":  "James",
			"Middle": "",
			"Last":   "Bond",
		},
		"Address": map[string]any{
			"Street":  "44",
			"City":    "London",
			"State":   "",
			"Country": "EN",
			"Zip":     "34-0595",
		},
	}

	eq, d = Eq(mA, mB)
	print(eq)
	print("\n")
	print(d)
	print("\n")

	tA := T{
		ID:   1,
		List: a,
		Name: &T2{
			First:  "Elmer",
			Middle: "James",
			Last:   "Fudd",
		},
		Address: T3{
			Street:  "123",
			City:    "Huntsville",
			State:   "AL",
			Country: "US",
			Zip:     "12345",
		},
	}
	tB := T{
		ID:   1,
		List: b,
		Name: &T2{
			First:  "James",
			Middle: "",
			Last:   "Bond",
		},
		Address: T3{
			Street:  "44",
			City:    "London",
			State:   "",
			Country: "EN",
			Zip:     "34-0595",
		},
	}

	eq, d = Eq(tA, tB)
	print(eq)
	print("\n")
	print(d)
	print("\n")
}

type T struct {
	ID      int
	List    []int
	Name    *T2
	Address T3
}

type T2 struct {
	First  string
	Middle string
	Last   string
	List   []int
}

type T3 struct {
	Street  string
	City    string
	State   string
	Country string
	Zip     string
}
