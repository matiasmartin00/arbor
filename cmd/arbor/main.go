package main

import (
	"fmt"
	"os"

	"github.com/matiasmartin00/arbor/cmd/arbor/cli"
)

func main() {
	rootCmd := cli.NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
