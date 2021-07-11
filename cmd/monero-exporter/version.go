package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "dev"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print the version of this CLI",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(version, commit)
	},
}
