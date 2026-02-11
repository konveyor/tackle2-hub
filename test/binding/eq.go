package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp/sort"
)

func init() {
	sort.Add(
		sort.ById,
		[]api.Ref{},
		[]*api.Ref{})
}
