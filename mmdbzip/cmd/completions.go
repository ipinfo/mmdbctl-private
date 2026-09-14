package cmd

import (
	"github.com/ipinfo/cli/lib/complete"
	"github.com/ipinfo/cli/lib/complete/predict"
)

// Completions returns the shell-completion tree for the mmdbzip commands.
//
// It is ready to mount as a subcommand of a larger CLI (e.g. under "zip" in
// mmdbctl) or to serve as a root. A fresh tree is built per call, so callers
// may add entries to Sub without affecting anyone else.
func Completions() *complete.Command {
	return &complete.Command{
		Sub: map[string]*complete.Command{
			"compress": completionsCompress,
			"verify":   completionsVerify,
			"analyze":  completionsAnalyze,
			"bench":    completionsBench,
		},
		Flags: map[string]complete.Predictor{
			"-h":     predict.Nothing,
			"--help": predict.Nothing,
		},
	}
}
