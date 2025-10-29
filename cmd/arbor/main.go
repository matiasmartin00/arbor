package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/matiasmartin00/arbor/internal/repo"
)

func main() {
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
	case "commit":
		if len(os.Args) < 4 || os.Args[2] != "-m" {
			fmt.Println("Usage: arbor commit -m <message>")
			os.Exit(1)
		}

		msg := strings.Join(os.Args[3:], " ")
		if err := repo.EnsureRepo("."); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		commitHash, err := repo.Commit(".", msg)
		if err != nil {
			fmt.Println("Error creating commit:", err)
			os.Exit(1)
		}

		fmt.Println("Commit created with hash:", commitHash)
	case "log":
		if err := repo.EnsureRepo("."); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		if err := repo.Log("."); err != nil {
			fmt.Println("Error displaying log:", err)
			os.Exit(1)
		}
	default:
		fmt.Println("Unknown command:", command)
		os.Exit(1)
	}
}
