# Arbor

Arbor is a minimal version control system written in Go.  
Its purpose is educational — to explore the inner mechanics of Git: how it stores objects, builds trees, manages commits, and tracks branches.

## Features
- Local repository structure (`.arbor/`)
- Object types: `blob`, `tree`, `commit`
- Simple staging area (index)
- References (`refs/heads`, `HEAD`)
- Basic commands: `init`, `add`, `commit`, `log`, `checkout`

## Usage

### Initialize a repository
```bash
arbor init
```

### Add a file
```bash
arbor add file.txt
```

### Create a commit
```bash
arbor commit -m "message"
```

### View commit history
```bash
arbor log
```