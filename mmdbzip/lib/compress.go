package lib

import "errors"

// Compress reads the mmdb file at inputPath, deduplicates identical subtrees of
// its search trie, and writes the smaller result to outputPath. The input file
// is left untouched.
func Compress(inputPath, outputPath string) error {
	return errors.New("not implemented")
}
