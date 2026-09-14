package cmd

import (
	"fmt"

	"github.com/ipinfo/cli/lib/complete"
	"github.com/ipinfo/cli/lib/complete/predict"
	"github.com/ipinfo/mmdbctl/mmdbzip/lib"
)

var completionsAnalyze = &complete.Command{
	Flags: map[string]complete.Predictor{
		"-h":     predict.Nothing,
		"--help": predict.Nothing,
	},
}

func printHelpAnalyze(prog string) {
	fmt.Printf(
		`Usage: %s analyze [<opts>] <mmdb_file>

Description:
  Estimate how much an mmdb file would shrink if compressed, without writing
  anything. Useful for deciding whether a dataset is worth compressing.

Options:
  General:
    --help, -h
      show help.
`, prog)
}

func cmdAnalyze(prog string, args []string) error {
	var help bool

	fs := newFlagSet("analyze")
	fs.BoolVarP(&help, "help", "h", false, "show help.")
	if err := fs.Parse(args); err != nil {
		return err
	}

	rest := fs.Args()
	if help || len(rest) == 0 {
		printHelpAnalyze(prog)
		return nil
	}

	return lib.Analyze(rest[0])
}
