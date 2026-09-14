package main

import (
	"os"

	zipcmd "github.com/ipinfo/mmdbctl/mmdbzip/cmd"
)

func cmdZip() error {
	return zipcmd.Execute(progBase+" zip", os.Args[2:])
}
