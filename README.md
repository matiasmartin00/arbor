# Arbor

Arbor is a minimal version control system written in Go.  
Its goal is educational — to understand how Git works internally: object storage, trees, commits, branching, merging, and working directory tracking.

## Features
- Local repository structure (`.arbor/`)
- Object types: `blob`, `tree`, `commit`
- Simple staging area (index)
- References (`refs/heads`, `HEAD`)
- Branch management
- Merge support (fast-forward and three-way)
- Working directory state tracking (`status`)
- Commands implemented so far:
  - `init`
  - `add`
  - `commit`
  - `log`
  - `checkout`
  - `branch`
  - `status`
  - `diff`
  - `merge`

## Usage

### Initialize a repository
```bash
arbor init
```

### Add files to the staging area
```bash
arbor add file1.txt file2.txt
arbor add .
arbor add *.txt **/*.txt
```

### Create a commit
```bash
arbor commit -m "your message"
arbor commit --message "your message"
```

### View commit history
```bash
arbor log
```

### Show file differences
Compare working directory and last commit:
```bash
arbor diff
```
Compare two commits:
```bash
arbor diff <commit1> <commit2>
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

### Merge branches
Fast-forward or three-way merge:
```bash
arbor merge <branch-name>
```
- If the branch is ahead of the current branch (fast-forward), Arbor updates the current branch reference and working directory.
- If branches diverged, Arbor performs a three-way merge and creates a merge commit.
- Conflicts are shown inline with conflict markers.

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
arbor commit -m "update on dev"
arbor checkout main
arbor merge dev
```

## Roadmap
Planned improvements and features:
- `tag` — lightweight and annotated tags  
- Colored console
- `remote`, `push`, `pull` — distributed synchronization  
