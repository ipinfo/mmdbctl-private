package cmd

import (
	"errors"
	"fmt"

	"github.com/ipinfo/cli/lib/complete"
	"github.com/ipinfo/cli/lib/complete/predict"
	"github.com/ipinfo/mmdbctl/mmdbzip/lib"
)

var completionsVerify = &complete.Command{
	Flags: map[string]complete.Predictor{
		"-h":     predict.Nothing,
		"--help": predict.Nothing,
	},
}

func printHelpVerify(prog string) {
	fmt.Printf(
		`Usage: %s verify [<opts>] <baseline_mmdb_file> <compressed_mmdb_file>

Description:
  Check that two mmdb files answer every lookup identically. Compares
  metadata, probes edge IPs, samples random lookups, and enumerates the full
  prefix set. Reports the first mismatch found.

Options:
  General:
    --help, -h
      show help.
`, prog)
}

func cmdVerify(prog string, args []string) error {
	var help bool

	fs := newFlagSet("verify")
	fs.BoolVarP(&help, "help", "h", false, "show help.")
	if err := fs.Parse(args); err != nil {
		return err
	}

	rest := fs.Args()
	if help || len(rest) == 0 {
		printHelpVerify(prog)
		return nil
	}
	if len(rest) < 2 {
		return errors.New("compressed mmdb file required as second argument")
	}

	return lib.Verify(rest[0], rest[1])
}
