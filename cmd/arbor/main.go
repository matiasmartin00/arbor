package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/matiasmartin00/arbor/cmd/arbor/cli"
	"github.com/matiasmartin00/arbor/internal/add"
	"github.com/matiasmartin00/arbor/internal/branch"
	"github.com/matiasmartin00/arbor/internal/checkout"
	"github.com/matiasmartin00/arbor/internal/commit"
	"github.com/matiasmartin00/arbor/internal/log"
	"github.com/matiasmartin00/arbor/internal/repo"
)

func main() {
	rootCmd := cli.NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func main_bkp() {
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

		added, err := add.Add(".", os.Args[2:])
		if err != nil {
			fmt.Println("Error adding file:", err)
			os.Exit(1)
		}

		for p, h := range added {
			fmt.Printf("Added %s with hash %s\n", p, h)
		}
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

		commitHash, err := commit.Commit(".", msg)
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

		if err := log.Log("."); err != nil {
			fmt.Println("Error displaying log:", err)
			os.Exit(1)
		}
	case "checkout":
		if len(os.Args) < 3 {
			fmt.Println("Usage: arbor checkout <commit-hash|branch-name>")
			os.Exit(1)
		}

		if err := repo.EnsureRepo("."); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		commitHashOrRef := os.Args[2]
		if err := checkout.Checkout(".", commitHashOrRef); err != nil {
			fmt.Println("Error during checkout:", err)
			os.Exit(1)
		}

		fmt.Println("Checked out to:", commitHashOrRef)
	case "branch":
		if len(os.Args) < 2 {
			fmt.Println("Usage: arbor branch <create|list> [args...]")
			os.Exit(1)
		}

		subcommand := os.Args[2]
		switch subcommand {
		case "create":
			if len(os.Args) < 4 {
				fmt.Println("Usage: arbor branch create <branch-name>")
				os.Exit(1)
			}

			if err := repo.EnsureRepo("."); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}

			branchName := os.Args[3]
			err := branch.CreateBranch(".", branchName)
			if err != nil {
				fmt.Println("Error creating branch:", err)
				os.Exit(1)
			}

			fmt.Println("Branch created:", branchName)
		case "list":
			if err := repo.EnsureRepo("."); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}

			branches, err := branch.ListBranches(".")
			if err != nil {
				fmt.Println("Error listing branches:", err)
				os.Exit(1)
			}

			fmt.Println("Branches:")
			for _, b := range branches {
				fmt.Println(" -", b)
			}
		default:
			fmt.Println("Unknown branch command:", subcommand)
			os.Exit(1)
		}
	default:
		fmt.Println("Unknown command:", command)
		os.Exit(1)
	}
}
