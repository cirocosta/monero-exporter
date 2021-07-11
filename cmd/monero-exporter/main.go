package main

import (
	"fmt"
	"os"
)

func main() {
	cmd := (&command{}).Cmd()
	cmd.AddCommand(versionCmd)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
