package cmd

import (
	"errors"
	"fmt"

	"github.com/ipinfo/cli/lib/complete"
	"github.com/ipinfo/cli/lib/complete/predict"
	"github.com/ipinfo/mmdbctl/mmdbzip/lib"
)

var completionsBenchPaired = &complete.Command{
	Flags: map[string]complete.Predictor{
		"-h":     predict.Nothing,
		"--help": predict.Nothing,
	},
}

func printHelpBenchPaired(prog string) {
	fmt.Printf(
		`Usage: %s bench paired [<opts>] <baseline_mmdb_file> <compressed_mmdb_file>

Description:
  Compare lookup throughput of two mmdb files. Runs many rounds, shuffling
  file order each round so both files see similar system load, then reports
  the paired delta with a confidence interval. Much less noisy than running
  two single-file benchmarks back to back.

Options:
  General:
    --help, -h
      show help.
`, prog)
}

func cmdBenchPaired(prog string, args []string) error {
	var help bool

	fs := newFlagSet("bench paired")
	fs.BoolVarP(&help, "help", "h", false, "show help.")
	if err := fs.Parse(args); err != nil {
		return err
	}

	rest := fs.Args()
	if help || len(rest) == 0 {
		printHelpBenchPaired(prog)
		return nil
	}
	if len(rest) < 2 {
		return errors.New("compressed mmdb file required as second argument")
	}

	return lib.BenchPaired(rest[0], rest[1])
}
