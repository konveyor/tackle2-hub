package seed

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestMerge(t *testing.T) {
	g := gomega.NewWithT(t)

	seeder := Target{}

	// the seed contains 10 targets in a given order, 3 of which are new
	seedOrder := []uint{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	// the user has set up a custom order for the 7 targets that already exist in the db
	userOrder := []uint{6, 9, 5, 4, 1, 3, 2}
	// the DB in total has 13 targets including the 3 newly seeded ones and 3 that were pre-existing
	// in the DB not not represented in the ordering due to a previous bug.
	allIds := []uint{11, 12, 13, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	// we expect the newly added targets to be woven into the user's custom ordering, with any targets
	// that had previously been dropped on the floor being added to the end of the ordering.
	expectedOrder := []uint{6, 7, 8, 9, 10, 5, 4, 1, 3, 2, 11, 12, 13}

	mergedOrder := seeder.merge(userOrder, seedOrder, allIds)
	g.Expect(mergedOrder).To(gomega.Equal(expectedOrder))
}
