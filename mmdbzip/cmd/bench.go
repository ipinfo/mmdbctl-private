package cmd

import (
	"fmt"

	"github.com/ipinfo/cli/lib/complete"
	"github.com/ipinfo/cli/lib/complete/predict"
	"github.com/ipinfo/mmdbctl/mmdbzip/lib"
)

var completionsBench = &complete.Command{
	Sub: map[string]*complete.Command{
		"paired": completionsBenchPaired,
		"memory": completionsBenchMemory,
	},
	Flags: map[string]complete.Predictor{
		"-h":     predict.Nothing,
		"--help": predict.Nothing,
	},
}

func printHelpBench(prog string) {
	fmt.Printf(
		`Usage: %s bench [<opts>] <mmdb_file>

Description:
  Measure lookup throughput of a single mmdb file over random IPs.

Subcommands:
  paired  compare lookup throughput of two mmdb files.
  memory  report reader memory and page-fault behaviour for one mmdb file.

Options:
  General:
    --help, -h
      show help.
`, prog)
}

// cmdBench dispatches the `bench` subcommands, falling through to the
// single-file benchmark when no subcommand is given.
//
// args is everything after the `bench` token, so this behaves identically no
// matter how deeply `bench` itself is nested.
func cmdBench(prog string, args []string) error {
	if len(args) > 0 {
		switch args[0] {
		case "paired":
			return cmdBenchPaired(prog, args[1:])
		case "memory":
			return cmdBenchMemory(prog, args[1:])
		}
	}

	var help bool

	fs := newFlagSet("bench")
	fs.BoolVarP(&help, "help", "h", false, "show help.")
	if err := fs.Parse(args); err != nil {
		return err
	}

	rest := fs.Args()
	if help || len(rest) == 0 {
		printHelpBench(prog)
		return nil
	}

	return lib.Bench(rest[0])
}
