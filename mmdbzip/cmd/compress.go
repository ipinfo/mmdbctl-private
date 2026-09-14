package cmd

import (
	"errors"
	"fmt"

	"github.com/ipinfo/cli/lib/complete"
	"github.com/ipinfo/cli/lib/complete/predict"
	"github.com/ipinfo/mmdbctl/mmdbzip/lib"
)

var completionsCompress = &complete.Command{
	Flags: map[string]complete.Predictor{
		"-h":     predict.Nothing,
		"--help": predict.Nothing,
	},
}

func printHelpCompress(prog string) {
	fmt.Printf(
		`Usage: %s compress [<opts>] <input_mmdb_file> <output_mmdb_file>

Description:
  Deduplicate identical subtrees of the search trie and write a smaller mmdb
  file with identical lookup semantics. The input file is left untouched.

Options:
  General:
    --help, -h
      show help.
`, prog)
}

func cmdCompress(prog string, args []string) error {
	var help bool

	fs := newFlagSet("compress")
	fs.BoolVarP(&help, "help", "h", false, "show help.")
	if err := fs.Parse(args); err != nil {
		return err
	}

	rest := fs.Args()
	if help || len(rest) == 0 {
		printHelpCompress(prog)
		return nil
	}
	if len(rest) < 2 {
		return errors.New("output mmdb file required as second argument")
	}

	return lib.Compress(rest[0], rest[1])
}
