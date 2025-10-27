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
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Usage: arbor add <file>")
			os.Exit(1)
		}
		if err := repo.EnsureRepo("."); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		hash, err := repo.Add(".", os.Args[2])
		if err != nil {
			fmt.Println("Error adding file:", err)
			os.Exit(1)
		}

		fmt.Println("File added with hash:", hash)
	default:
		fmt.Println("Unknown command:", command)
	}
}
