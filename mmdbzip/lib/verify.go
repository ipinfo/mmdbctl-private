package lib

import "errors"

// Verify checks that two mmdb files answer every lookup identically, comparing
// metadata, edge IPs, sampled random lookups and the full prefix set.
func Verify(baselinePath, comparePath string) error {
	return errors.New("not implemented")
}
