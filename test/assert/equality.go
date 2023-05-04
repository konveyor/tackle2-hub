package assert

import (
	"fmt"
)

//
// Simple equality check working for flat types (no nested types passed by reference).
func FlatEqual(got, expected interface{}) bool {
	return fmt.Sprintf("%v", got) == fmt.Sprintf("%v", expected)
}
