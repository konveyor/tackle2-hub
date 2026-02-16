package cmp

import (
	"testing"
	"time"

	sort "github.com/konveyor/tackle2-hub/test/cmp/sort"
	. "github.com/onsi/gomega"
)

func TestEq(t *testing.T) {
	g := NewGomegaWithT(t)

	listA := []int{1, 2, 3}
	listB := []int{1, 3, 2}

	structA := T{
		ID:      1,
		Created: time.Now(),
		List:    listA,
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

	structB := T{
		ID:   1,
		List: listB,
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

	structC := T{
		ID:   1,
		List: listA,
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
			Zip:     "12345",
		},
	}

	mapA := map[string]any{
		"ID":   1,
		"List": listA,
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

	mapB := map[string]any{
		"ID":   1,
		"List": listB,
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

	testCases := []struct {
		id     int
		name   string
		a, b   any
		ignore []string
		wantEq bool
	}{
		// ── Strings ───────────────────────────────────────────────
		{name: "strings equal", a: "xx", b: "xx", wantEq: true},
		{name: "strings not equal", a: "hello", b: "world", wantEq: false},

		// ── Primitives ────────────────────────────────────────────
		{name: "int equal", a: 42, b: 42, wantEq: true},
		{name: "int not equal", a: 42, b: 100, wantEq: false},
		{name: "float equal", a: 3.14, b: 3.14, wantEq: true},
		{name: "float not equal", a: 3.14, b: 2.71, wantEq: false},
		{name: "bool equal", a: true, b: true, wantEq: true},
		{name: "bool not equal", a: true, b: false, wantEq: false},

		// ── Type mismatches ───────────────────────────────────────
		{name: "type mismatch (int vs string)", a: 42, b: "42", wantEq: false},
		{name: "type mismatch (int vs float)", a: 42, b: 42.0, wantEq: false},

		// ── Slices ────────────────────────────────────────────────
		{name: "[]int equal", a: listA, b: listA, wantEq: true},
		{name: "[]int not equal (order)", a: listA, b: listB, wantEq: false},
		{name: "[]int not equal (different values)", a: []int{1, 2, 3}, b: []int{1, 2, 4}, wantEq: false},
		{name: "[]int not equal (different length - longer first)", a: []int{1, 2, 3, 4}, b: listB, wantEq: false},
		{name: "[]int not equal (different length - shorter first)", a: listB, b: []int{1, 2, 3, 4, 5}, wantEq: false},
		{name: "empty slices equal", a: []int{}, b: []int{}, wantEq: true},
		{name: "nil slice vs empty slice", a: []int(nil), b: []int{}, wantEq: true},
		{name: "nil slice vs nil slice", a: []int(nil), b: []int(nil), wantEq: true},
		{name: "[]T6]", a: []T6{{Id: 1}, {Id: 2}}, b: []T6{{Id: 1}, {Id: 2}}, wantEq: true},
		{
			name:   "[]T6] name ignored.",
			a:      []T6{{Id: 1, Name: "xx"}, {Id: 2}},
			b:      []T6{{Id: 1}, {Id: 2}},
			wantEq: true,
			ignore: []string{".Name"}},

		// ── Nil ───────────────────────────────────────────────────
		{name: "nil vs non-nil", a: nil, b: 10, wantEq: false},
		{name: "non-nil vs nil", a: 10, b: nil, wantEq: false},
		{name: "nil vs nil", a: nil, b: nil, wantEq: true},

		// ── Pointers ──────────────────────────────────────────────
		{name: "pointer equal (same content)", a: &T2{First: "John"}, b: &T2{First: "John"}, wantEq: true},
		{name: "pointer not equal (different content)", a: &T2{First: "John"}, b: &T2{First: "Jane"}, wantEq: false},
		{name: "pointer vs nil", a: &T2{First: "John"}, b: (*T2)(nil), wantEq: false},
		{name: "nil pointer vs nil pointer", a: (*T2)(nil), b: (*T2)(nil), wantEq: true},

		// ── Maps ──────────────────────────────────────────────────
		{name: "map equal (same value)", a: mapA, b: mapA, wantEq: true},
		{name: "map not equal (different values)", a: mapA, b: mapB, wantEq: false},
		{name: "map with extra key in A", a: map[string]int{"a": 1, "b": 2}, b: map[string]int{"a": 1}, wantEq: false},
		{name: "map with extra key in B", a: map[string]int{"a": 1}, b: map[string]int{"a": 1, "b": 2}, wantEq: false},
		{name: "empty maps equal", a: map[string]int{}, b: map[string]int{}, wantEq: true},
		{name: "nil map vs nil map", a: map[string]int(nil), b: map[string]int(nil), wantEq: true},

		// ── Time ──────────────────────────────────────────────────
		{name: "time equal", a: time.Unix(1000, 0), b: time.Unix(1000, 0), wantEq: true},
		{name: "time not equal", a: time.Unix(1000, 0), b: time.Unix(2000, 0), wantEq: false},

		// ── Structs ───────────────────────────────────────────────
		{name: "struct equal (identical)", a: structA, b: structA, wantEq: true},
		{name: "struct not equal (different values)", a: structA, b: structB, wantEq: false},
		{
			name:   "struct equal when ignoring Created and Name.Born",
			a:      structA,
			b:      structC,
			ignore: []string{"Created", "Name.Born"},
			wantEq: true,
		},
		{
			name:   "struct equal when ignoring Created and entire Name",
			a:      structA,
			b:      structC,
			ignore: []string{"Created", "Name"},
			wantEq: true,
		},
		{
			name:   "struct with unexported fields",
			a:      T4{Public: "same", private: "different1"},
			b:      T4{Public: "same", private: "different2"},
			wantEq: true,
		},
		{
			name:   "struct with anonymous field",
			a:      T5{T2: T2{First: "John"}, Age: 30},
			b:      T5{T2: T2{First: "John"}, Age: 30},
			wantEq: true,
		},
		{
			name:   "struct with anonymous field and ignored Refs name.",
			a:      T5{T2: T2{First: "John"}, Age: 30, Refs: []T6{{Id: 1}, {Id: 2}}},
			b:      T5{T2: T2{First: "John"}, Age: 30, Refs: []T6{{Id: 1, Name: "xx"}, {Id: 2}}},
			ignore: []string{"Refs.Name"},
			wantEq: true,
		},
		{
			name:   "struct with anonymous field not equal",
			a:      T5{T2: T2{First: "John"}, Age: 30},
			b:      T5{T2: T2{First: "Jane"}, Age: 30},
			wantEq: false,
		},
		{
			name:   "struct ignoring non-existent path",
			a:      structA,
			b:      structB,
			ignore: []string{"NonExistent.Field"},
			wantEq: false,
		},
		{
			name: "struct with unsorted list",
			a: T8{
				List: []T7{
					{ID: 1, Name: "one"},
					{ID: 2, Name: "two"},
					{ID: 3, Name: "three"},
				},
			},
			b: T8{
				List: []T7{
					{ID: 3, Name: "three"},
					{ID: 1, Name: "one"},
					{ID: 2, Name: "two"},
				},
			},
			wantEq: true,
		},
	}

	sort.Add(sort.ById, []T7{})
	t.Cleanup(func() {
		sort.Reset()
	})
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			eq, report := Eq(tc.a, tc.b, tc.ignore...)
			if !eq {
				print(report)
			}
			g.Expect(eq).To(Equal(tc.wantEq))
			if eq != tc.wantEq {
				t.Logf("report:\n%s", report)
			}
		})
	}

	sort.Reset()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			eq, report := New().
				Sort(sort.ById, []T7{}).
				Ignore(tc.ignore...).
				Eq(tc.a, tc.b)
			if !eq {
				print(report)
			}
			g.Expect(eq).To(Equal(tc.wantEq))
			if eq != tc.wantEq {
				t.Logf("report:\n%s", report)
			}
		})
	}
}

func TestEqSort(t *testing.T) {
	g := NewGomegaWithT(t)

	a := []T7{
		{ID: 1, Name: "one"},
		{ID: 2, Name: "two"},
		{ID: 3, Name: "three"},
	}
	b := []T7{
		{ID: 3, Name: "three"},
		{ID: 1, Name: "one"},
		{ID: 2, Name: "two"},
	}
	cmp := New()
	cmp = cmp.Sort(sort.ById, []T7{})
	eq, report := cmp.Eq(a, b)
	g.Expect(eq).To(BeTrue(), report)
}

func TestEqReport(t *testing.T) {
	g := NewGomegaWithT(t)

	listA := []int{1, 2, 3}
	listB := []int{1, 3, 2}

	testCases := []struct {
		name         string
		a, b         any
		wantEq       bool
		wantInReport []string
	}{
		{
			name:         "strings not equal",
			a:            "hello",
			b:            "world",
			wantEq:       false,
			wantInReport: []string{`~  = "world" expected: "hello"`},
		},
		{
			name:         "int not equal",
			a:            42,
			b:            100,
			wantEq:       false,
			wantInReport: []string{"~  = 100 expected: 42"},
		},
		{
			name:   "type mismatch (int vs string)",
			a:      42,
			b:      "42",
			wantEq: false,
			wantInReport: []string{
				": (type) int != string",
			},
		},
		{
			name:   "[]int not equal (order)",
			a:      listA,
			b:      listB,
			wantEq: false,
			wantInReport: []string{
				"~ [1] = 3 expected: 2",
				"~ [2] = 2 expected: 3",
			},
		},
		{
			name: "struct not equal (different values)",
			a: T{
				ID:   1,
				List: listA,
				Name: &T2{
					First:  "Elmer",
					Middle: "James",
					Last:   "Fudd",
				},
			},
			b: T{
				ID:   1,
				List: listB,
				Name: &T2{
					First:  "James",
					Middle: "",
					Last:   "Bond",
				},
			},
			wantEq: false,
			wantInReport: []string{
				"~ List[1] = 3 expected: 2",
				`~ Name.First = "James" expected: "Elmer"`,
				`~ Name.Middle = "" expected: "James"`,
				`~ Name.Last = "Bond" expected: "Fudd"`,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			eq, report := Eq(tc.a, tc.b)
			g.Expect(eq).To(Equal(tc.wantEq))
			if tc.wantEq {
				g.Expect(report).To(BeEmpty())
			} else {
				g.Expect(report).ToNot(BeEmpty())
				for _, expected := range tc.wantInReport {
					g.Expect(report).To(ContainSubstring(expected))
				}
			}
		})
	}
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

type T4 struct {
	Public  string
	private string
}

type T5 struct {
	T2
	Age  int
	Refs []T6
}

type T6 struct {
	Id   int
	Name string
}

type T7 struct {
	ID   uint
	Name string
}

type T8 struct {
	List []T7
}
