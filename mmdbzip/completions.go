package main

import (
	"github.com/ipinfo/mmdbctl/mmdbzip/cmd"
)

func handleCompletions() {
	// The zip commands are the whole of this binary, so their tree is the
	// root; `completion` is this binary's own and gets added on top.
	completions := cmd.Completions()
	completions.Sub["completion"] = completionsCompletion

	completions.Complete(progBase)
}
