package lib

import "errors"

// BenchPaired compares lookup throughput of two mmdb files over many rounds,
// reporting the paired delta with a confidence interval.
func BenchPaired(baselinePath, comparePath string) error {
	return errors.New("not implemented")
}
