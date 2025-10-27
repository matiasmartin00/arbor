package main

import (
	"fmt"
	"os"

	"github.com/matiasmartin00/arbor/internal/repo"
)

func main() {
	fmt.Println("Arbor VCS - Command Line Interface")
	if len(os.Args) < 2 {
		fmt.Println("Usage: arbor <command> [options]")
		return
	}

	command := os.Args[1]
	switch command {
	case "init":
		err := repo.Init(".")
		if err != nil {
			fmt.Println("Error initializing repository:", err)
			os.Exit(1)
		}
		fmt.Println("Repository initialized.")
	default:
		fmt.Println("Unknown command:", command)
	}
}
