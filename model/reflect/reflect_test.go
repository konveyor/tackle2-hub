package reflect

import (
	"errors"
	"testing"

	"github.com/onsi/gomega"
)

func TestHasField(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	type B struct {
		Name string
		Age  string
	}
	type B2 struct {
		Name2 string
		Age2  string
	}
	type M struct {
		B
		*B2
		Ptr    *B
		Object B
		Int    int
		IntPtr *int
		List   []string
	}

	// Test expected.
	_, err := HasFields(
		&M{B2: &B2{}},
		"Name",
		"Age",
		"Name2",
		"Age2",
		"Ptr",
		"Object",
		"Int",
		"IntPtr",
		"List")
	g.Expect(err).To(gomega.BeNil())

	// Test anonymous NIL pointer.
	_, err = HasFields(
		&M{}, // PROBLEM HERE.
		"Name",
		"Age",
		"Name2",
		"Age2",
		"Ptr",
		"Object",
		"Int",
		"IntPtr",
		"List")
	g.Expect(err).ToNot(gomega.BeNil())

	// Invalid field.
	_, err = HasFields(
		&M{B2: &B2{}},
		"Name",
		"Age",
		"Name2",
		"Age2",
		"Ptr",
		"NOT-VALID", // PROBLEM HERE
		"Object",
		"Int",
		"IntPtr",
		"List")
	g.Expect(errors.Is(err, &FieldNotValid{})).To(gomega.BeTrue())
}
