# Arbor

Arbor is a minimal version control system written in Go.  
Its goal is educational — to understand how Git works internally: object storage, trees, commits, branching, and working directory tracking.

## Features
- Local repository structure (`.arbor/`)
- Object types: `blob`, `tree`, `commit`
- Simple staging area (index)
- References (`refs/heads`, `HEAD`)
- Branch management
- Working directory state tracking (`status`)
- Commands implemented so far:
  - `init`
  - `add`
  - `commit`
  - `log`
  - `checkout`
  - `branch`
  - `status`

## Usage

### Initialize a repository
```bash
arbor init
```

### Add files to the staging area
```bash
arbor add file1.txt file2.txt
```

### Create a commit
```bash
arbor commit -m "your message"
```

### View commit history
```bash
arbor log
```

### Switch branches or commits
```bash
arbor checkout <branch-name | commit-hash>
```

### Manage branches
Create a new branch:
```bash
arbor branch create feature-xyz
```

List all branches:
```bash
arbor branch list
```

### Check repository status
```bash
arbor status
```
Displays:
- Changes to be committed (staged)
- Changes not staged for commit (modified in working directory)
- Untracked files

## Example workflow
```bash
arbor init
echo "hello" > file.txt
arbor add file.txt
arbor commit -m "initial commit"
arbor branch create dev
arbor checkout dev
echo "change" >> file.txt
arbor status
```

## Roadmap
Planned features:
- `diff` — show file differences between commits, index, and working directory  
- `merge` — merge branches and handle conflicts  
- `tag` — lightweight and annotated tags  
- `remote`, `push`, `pull` — distributed synchronization
