package client

import "fmt"

func FlatEqual(got, expected interface{}) bool {
	return fmt.Sprintf("%v", got) == fmt.Sprintf("%v", expected)
}
