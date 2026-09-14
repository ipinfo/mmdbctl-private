// Package cmd is the command-line surface for mmdbzip: flag parsing, help
// text and subcommand dispatch. The compression logic itself lives in
// github.com/ipinfo/mmdbctl/mmdbzip/lib, which knows nothing about the CLI.
//
// Every entry point takes the program prefix to show in usage strings, so the
// same commands serve both the standalone `mmdbzip` binary and `mmdbctl zip`.
package cmd

import (
	"io"

	"github.com/spf13/pflag"
)

// newFlagSet returns a flag set for a single (sub)command.
//
// Parse errors are reported through the error it returns rather than printed
// by the flag set itself, so callers surface them like any other command error
// and keep control of their own help output.
func newFlagSet(name string) *pflag.FlagSet {
	fs := pflag.NewFlagSet(name, pflag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

// Execute runs a mmdbzip command.
//
// prog is the prefix shown in usage strings, e.g. "mmdbzip" for the standalone
// binary or "mmdbctl zip" when nested. args is everything after that prefix,
// so callers strip exactly the tokens they own and no layer needs to know how
// deeply it is nested.
func Execute(prog string, args []string) error {
	var cmd string
	if len(args) > 0 {
		cmd = args[0]
	}

	// Everything after the command name.
	var rest []string
	if len(args) > 1 {
		rest = args[1:]
	}

	switch cmd {
	case "compress":
		return cmdCompress(prog, rest)
	case "verify":
		return cmdVerify(prog, rest)
	case "analyze":
		return cmdAnalyze(prog, rest)
	case "bench":
		return cmdBench(prog, rest)
	default:
		return cmdDefault(prog, args)
	}
}
