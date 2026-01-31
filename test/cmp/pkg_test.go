package cmp

/*
import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestCmp(t *testing.T) {
	g := NewGomegaWithT(t)

	//
	// strings eq.
	eq, d := EQ("xx", "xx")
	if !eq {
		print("eq=")
		print(eq)
		print("\n")
		print(d)
	}
	g.Expect(eq).To(BeTrue())
	// strings not eq.
	eq, d = EQ("hello", "world")
	if eq {
		print("eq=")
		print(eq)
		print("\n")
		print(d)
	}
	g.Expect(eq).To(BeFalse())

	//
	// []int eq.
	a := []int{1, 2, 3}
	b := []int{1, 2, 3}
	eq, d = EQ(a, b)
	if !eq {
		print(eq)
		print("\n")
		print(d)
	}
	g.Expect(eq).To(BeTrue())
	// []int not eq.
	a = []int{1, 2, 3}
	b = []int{1, 3, 2}
	eq, d = EQ(a, b)
	if eq {
		print(eq)
		print("\n")
		print(d)
	}
	g.Expect(eq).To(BeFalse())
	// []int not eq (length)
	a = []int{1, 2, 3, 4}
	b = []int{1, 3, 2}
	eq, d = EQ(a, b)
	g.Expect(eq).To(BeFalse())
	print(eq)
	print("\n")
	print(d)
	print("\n")
	// []int not eq (length)
	a = []int{1, 3, 2}
	b = []int{1, 2, 3, 4, 5}
	eq, d = EQ(a, b)
	g.Expect(eq).To(BeFalse())
	print(eq)
	print("\n")
	print(d)
	print("\n")

	//
	// NIL not eq.
	eq, d = EQ(nil, 10)
	g.Expect(eq).To(BeFalse())
	print(eq)
	print("\n")
	print(d)
	print("\n")

	//
	// Map eq.
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
	mB := mA
	eq, d = EQ(mA, mB)
	g.Expect(eq).To(BeTrue())
	// Map not eq.
	mA = map[string]any{
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
	mB = map[string]any{
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
	eq, d = EQ(mA, mB)
	g.Expect(eq).To(BeFalse())
	print(eq)
	print("\n")
	print(d)
	print("\n")

	// Struct eq.
	tA := T{
		ID:      1,
		Created: time.Now(),
		List:    a,
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
	tB := tA
	eq, d = EQ(tA, tB)
	g.Expect(eq).To(BeTrue())
	// Struct not eq.
	tA = T{
		ID:      1,
		Created: time.Now(),
		List:    a,
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
	tB = T{
		ID:   1,
		List: b,
		Name: &T2{
			First:  "James",
			Middle: "",
			Last:   "Bond",
			Born:   time.Now(),
		},
		Address: T3{
			Street:  "44",
			City:    "London",
			State:   "",
			Country: "EN",
			Zip:     "34-0595",
		},
	}
	tC := T{
		ID:   1,
		List: a,
		Name: &T2{
			First:  "Elmer",
			Middle: "James",
			Last:   "Fudd",
			Born:   time.Now(),
		},
		Address: T3{
			Street:  "123",
			City:    "Huntsville",
			State:   "AL",
			Country: "US",
			Zip:     "123456",
		},
	}
	eq, d = EQ(tA, tB)
	g.Expect(eq).To(BeFalse())
	print(eq)
	print("\n")
	print(d)
	print("\n")
	// Struct eq Created, Born ignored.
	eq, d = EQ(tA, tC, "Created", "Name.Born")
	print(eq)
	print("\n")
	print(d)
	print("\n")
	g.Expect(eq).To(BeTrue())
	// Struct eq Created, Name.* ignored.
	eq, d = EQ(tA, tC, "Created", "Name")
	print(eq)
	print("\n")
	print(d)
	print("\n")
	g.Expect(eq).To(BeTrue())
}

type T struct {
	ID      int
	Created time.Time
	List    []int
	Name    *T2
	Address T3
}

type T2 struct {
	First  string
	Middle string
	Last   string
	List   []int
	Born   time.Time
}

type T3 struct {
	Street  string
	City    string
	State   string
	Country string
	Zip     string
}
*/
