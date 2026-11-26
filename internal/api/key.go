package api

import (
	"fmt"
	"net/url"
)

// validKey returns an error when the key (setting key, fact source or fact name)
// contains characters that must be percent-encoded in a URL path segment.
// The settings and facts sub-resources are addressed by key rather than by ID,
// so a key such as 'my key' or 'a/b' could be created but not reached again.
func validKey(kind, key string) (err error) {
	if url.PathEscape(key) != key {
		err = &BadRequestError{
			Reason: fmt.Sprintf(
				"%s %q contains characters that require URL encoding.",
				kind,
				key),
		}
	}
	return
}
