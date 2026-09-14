package cmd

import (
	"fmt"

	"github.com/ipinfo/cli/lib/complete"
	"github.com/ipinfo/cli/lib/complete/predict"
	"github.com/ipinfo/mmdbctl/mmdbzip/lib"
)

var completionsBenchMemory = &complete.Command{
	Flags: map[string]complete.Predictor{
		"-h":     predict.Nothing,
		"--help": predict.Nothing,
	},
}

func printHelpBenchMemory(prog string) {
	fmt.Printf(
		`Usage: %s bench memory [<opts>] <mmdb_file>

Description:
  Report how much memory a reader actually uses for an mmdb file after a
  lookup workload, along with its page-fault behaviour. File size on disk is
  not the same as resident memory, so this is the tool for answering whether
  a database fits a given memory budget.

Options:
  General:
    --help, -h
      show help.
`, prog)
}

func cmdBenchMemory(prog string, args []string) error {
	var help bool

	fs := newFlagSet("bench memory")
	fs.BoolVarP(&help, "help", "h", false, "show help.")
	if err := fs.Parse(args); err != nil {
		return err
	}

	rest := fs.Args()
	if help || len(rest) == 0 {
		printHelpBenchMemory(prog)
		return nil
	}

	return lib.BenchMemory(rest[0])
}
