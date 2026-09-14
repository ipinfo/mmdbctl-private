package cmd

import (
	"fmt"
)

func printHelpDefault(prog string) {
	fmt.Printf(
		`Usage: %s <cmd> [<opts>] [<args>]

Description:
  Losslessly compress mmdb files by deduplicating identical subtrees of the
  search trie. The output is a standard mmdb file with identical lookup
  semantics that any reader can open.

Commands:
  compress  compress an mmdb file.
  verify    check that two mmdb files answer lookups identically.
  analyze   estimate how much an mmdb file would compress.
  bench     measure lookup throughput of an mmdb file.

Options:
  General:
    --help, -h
      show help.
`, prog)
}

func cmdDefault(prog string, args []string) error {
	var help bool

	fs := newFlagSet(prog)
	fs.BoolVarP(&help, "help", "h", false, "show help.")
	if err := fs.Parse(args); err != nil {
		return err
	}

	// currently we do nothing by default.
	printHelpDefault(prog)
	return nil
}
