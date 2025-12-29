# System Commands Reference

This project is developed on **macOS/Linux**. Standard Unix commands are available.

## File System Operations

```bash
# List files
ls -la                    # Detailed list with hidden files
ls -lh                    # Human-readable sizes
ls -lt                    # Sort by modification time

# List directory tree
tree -L 3                 # Limit depth to 3 levels
tree -I 'node_modules'    # Ignore directories

# Change directory
cd path/to/directory
cd ..                     # Parent directory
cd ~                      # Home directory
cd -                      # Previous directory
pwd                       # Print working directory

# Create directory
mkdir dirname
mkdir -p path/to/nested/dir

# Remove files/directories
rm file.txt
rm -rf directory/         # Recursive, force
rm -i file.txt           # Interactive (ask confirmation)

# Copy files
cp source.txt destination.txt
cp -r source_dir/ dest_dir/
cp -p file.txt backup.txt  # Preserve attributes

# Move/rename files
mv old_name.txt new_name.txt
mv file.txt directory/

# Find files
find . -name "*.go"
find . -type f -name "*.proto"
find . -type d -name "vendor"
find . -mtime -7          # Modified in last 7 days
find . -size +1M          # Files larger than 1MB

# Check file/directory existence
test -f file.txt && echo "file exists"
test -d directory && echo "directory exists"
[ -f file.txt ] && echo "exists"

# Create symbolic links
ln -s /path/to/original /path/to/link

# File permissions and info
chmod +x script.sh        # Make executable
chmod 644 file.txt        # rw-r--r--
chmod 755 script.sh       # rwxr-xr-x
ls -l file.txt           # View permissions
stat file.txt            # Detailed file info
```

## Text Processing

```bash
# Search in files (grep)
grep -r "pattern" .
grep -rn "pattern" .              # With line numbers
grep -ri "pattern" .              # Case insensitive
grep -rn "pattern" --include="*.go" .
grep -v "pattern" file.txt        # Invert match (exclude)
grep -E "regex_pattern" file.txt  # Extended regex
grep -A 3 "pattern" file.txt      # 3 lines after match
grep -B 3 "pattern" file.txt      # 3 lines before match
grep -C 3 "pattern" file.txt      # 3 lines context

# View file contents
cat file.txt
less file.txt            # Paginated view (q to quit)
more file.txt            # Simple paginated view
head -n 20 file.txt      # First 20 lines
tail -n 20 file.txt      # Last 20 lines
tail -f file.txt         # Follow (watch for changes)

# Count lines/words/chars
wc -l file.txt           # Count lines
wc -w file.txt           # Count words
wc -c file.txt           # Count bytes

# Text manipulation
sed 's/old/new/g' file.txt              # Replace (all occurrences)
sed 's/old/new/' file.txt               # Replace (first occurrence per line)
sed -i '' 's/old/new/g' file.txt        # In-place edit (macOS)
sed -i 's/old/new/g' file.txt           # In-place edit (Linux)
awk '{print $1}' file.txt               # Print first column
awk -F',' '{print $2}' file.csv         # CSV - print 2nd column
cut -d',' -f2 file.csv                  # CSV - cut 2nd field
sort file.txt                           # Sort lines
sort -r file.txt                        # Reverse sort
sort -u file.txt                        # Sort and remove duplicates
uniq file.txt                           # Remove adjacent duplicates
tr 'a-z' 'A-Z' < file.txt              # Translate to uppercase
```

## Process Management

```bash
# List processes
ps aux
ps aux | grep service-name
ps -ef                   # Full format listing
top                      # Interactive process viewer (q to quit)
htop                     # Better top (if installed)

# Kill process
kill PID
kill -9 PID              # Force kill
kill -15 PID             # Graceful shutdown (SIGTERM)
killall process-name     # Kill by name
pkill pattern            # Kill by pattern

# Background/foreground jobs
command &                # Run in background
jobs                     # List background jobs
fg %1                    # Bring job 1 to foreground
bg %1                    # Resume job 1 in background
Ctrl+Z                   # Suspend current process
disown %1                # Detach job from shell

# Process priority
nice -n 10 command       # Run with lower priority
renice -n 5 -p PID       # Change priority of running process
```

## Network Operations

```bash
# Check port usage
lsof -i :8000            # macOS/Linux
netstat -tlnp | grep :8000  # Linux
lsof -i -P               # All listening ports

# HTTP requests (curl)
curl http://localhost:8000/health
curl -i http://localhost:8000/health        # Include headers
curl -X POST http://localhost:8000/v1/symbols -d '{"name":"test"}' -H "Content-Type: application/json"
curl -X POST http://localhost:8000/v1/symbols -d @request.json
curl -o output.txt http://example.com/file  # Save to file
curl -L http://example.com                  # Follow redirects
curl -v http://example.com                  # Verbose

# DNS lookup
nslookup domain.com
dig domain.com
dig +short domain.com
host domain.com

# Network connectivity
ping google.com
ping -c 4 google.com     # Ping 4 times
traceroute google.com    # Trace route
nc -zv localhost 8000    # Test port connectivity
```

## Git Operations

```bash
# Status and history
git status
git status -s            # Short format
git log
git log --oneline
git log --oneline -10
git log --graph --oneline --all
git diff
git diff HEAD~1          # Diff with previous commit
git diff --staged        # Diff staged changes
git show commit-hash     # Show commit details

# Branching
git branch               # List branches
git branch -a            # List all branches (including remote)
git checkout -b feature-branch
git checkout main
git branch -d branch-name           # Delete branch
git branch -D branch-name           # Force delete

# Staging and committing
git add .
git add file.txt
git add -p               # Interactive staging
git reset HEAD file.txt  # Unstage file
git commit -m "message"
git commit --amend       # Amend last commit

# Remote operations
git push origin branch-name
git push -u origin branch-name      # Set upstream
git pull origin main
git pull --rebase origin main
git fetch
git fetch --all
git remote -v            # List remotes
git remote add origin url

# Stashing
git stash
git stash save "message"
git stash list
git stash pop
git stash apply
git stash drop
git stash clear

# Undoing changes
git checkout -- file.txt            # Discard changes
git reset --soft HEAD~1             # Undo last commit, keep changes
git reset --hard HEAD~1             # Undo last commit, discard changes
git revert commit-hash              # Create revert commit

# Tagging
git tag v1.0.0
git tag -a v1.0.0 -m "version 1.0.0"
git push origin v1.0.0
git push origin --tags
```

## Docker Operations

```bash
# Container management
docker ps                           # List running containers
docker ps -a                        # List all containers
docker stop container-name
docker start container-name
docker restart container-name
docker rm container-name            # Remove container
docker rm -f container-name         # Force remove

# Logs
docker logs container-name
docker logs -f container-name       # Follow logs
docker logs --tail 100 container-name
docker logs --since 1h container-name

# Docker Compose
docker-compose up
docker-compose up -d                # Detached mode
docker-compose down
docker-compose down -v              # Remove volumes
docker-compose logs service-name
docker-compose logs -f service-name
docker-compose ps
docker-compose restart service-name
docker-compose build                # Rebuild images
docker-compose up --build           # Rebuild and start

# Execute in container
docker exec -it container-name bash
docker exec -it container-name sh
docker exec container-name command

# Images
docker images
docker rmi image-name
docker pull image-name
docker build -t tag-name .
docker system prune                 # Clean up

# Inspect
docker inspect container-name
docker stats                        # Resource usage
docker top container-name           # Running processes
```

## Go Commands

```bash
# Module operations
go mod init module-name
go mod tidy                         # Add missing, remove unused
go mod download
go mod verify
go mod why package-name
go mod graph

# Building
go build ./...
go build -o output_name ./cmd/service
go build -v ./...                   # Verbose
go install ./cmd/service

# Testing
go test ./...
go test -v ./internal/biz/
go test -run TestName
go test -run TestName/SubtestName
go test -race ./...                 # Race detection
go test -cover ./...                # Coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
go test -bench=. ./...              # Benchmarks
go test -timeout 30s ./...

# Running
go run ./cmd/service/
go run ./cmd/service/main.go

# Formatting and linting
gofmt -s -w .
go fmt ./...
go vet ./...
golangci-lint run                   # If installed

# Code generation
go generate ./...

# Dependencies
go get package-name
go get package-name@version
go get -u package-name              # Update
go list ./...
go list -m all                      # List modules
go list -m -versions package-name   # List versions

# Workspace operations (Go 1.18+)
go work init
go work use ./module1 ./module2
go work edit -print
go work sync
```

## Environment Variables

```bash
# View environment
env
printenv
echo $PATH
echo $GOPATH

# Set variable (current session)
export VAR_NAME=value
export GOPATH=$HOME/go

# Unset variable
unset VAR_NAME

# Set for single command
VAR_NAME=value command

# Check Go environment
go env
go env GOPATH
go env GOROOT
go env GOPROXY
```

## Disk Usage

```bash
# Check disk space
df -h
df -h /path/to/directory

# Check directory size
du -sh directory/
du -h --max-depth=1
du -sh *                            # Size of each item in current dir

# Find large files
find . -type f -size +100M
du -ah . | sort -rh | head -20      # 20 largest files/dirs
```

## System Information

```bash
# OS information
uname -a
uname -s                            # Kernel name
uname -r                            # Kernel release
sw_vers                             # macOS version
cat /etc/os-release                 # Linux version

# CPU information
sysctl -n machdep.cpu.brand_string  # macOS
lscpu                               # Linux
cat /proc/cpuinfo                   # Linux

# Memory information
top -l 1 -s 0 | grep PhysMem        # macOS
free -h                             # Linux
cat /proc/meminfo                   # Linux

# Hostname
hostname
hostname -f                         # Fully qualified
```

## Compression and Archives

```bash
# tar archives
tar -czf archive.tar.gz directory/  # Create compressed archive
tar -xzf archive.tar.gz             # Extract compressed archive
tar -tzf archive.tar.gz             # List contents
tar -xzf archive.tar.gz -C /path/   # Extract to path

# zip
zip -r archive.zip directory/
unzip archive.zip
unzip -l archive.zip                # List contents
unzip archive.zip -d /path/         # Extract to path

# gzip
gzip file.txt                       # Compress
gunzip file.txt.gz                  # Decompress
```

## Useful Shortcuts and Tricks

```bash
# Command history
history
history | grep command
!123                                # Run command number 123
!!                                  # Run last command
!$                                  # Last argument of previous command
sudo !!                             # Run last command with sudo
Ctrl+R                              # Reverse search history

# Pipes and redirection
command1 | command2                 # Pipe output
command > file.txt                  # Redirect output (overwrite)
command >> file.txt                 # Redirect output (append)
command 2>&1                        # Redirect stderr to stdout
command > /dev/null 2>&1            # Discard all output
command < input.txt                 # Redirect input
command1 && command2                # Run command2 if command1 succeeds
command1 || command2                # Run command2 if command1 fails
command1 ; command2                 # Run both regardless

# Job control
Ctrl+C                              # Interrupt (SIGINT)
Ctrl+Z                              # Suspend
Ctrl+D                              # EOF / Exit shell

# Aliases (add to ~/.bashrc or ~/.zshrc)
alias ll='ls -lah'
alias ..='cd ..'
alias gs='git status'
alias gp='git pull'

# Quick file editing
nano file.txt                       # Simple editor
vi file.txt                         # Vi editor
vim file.txt                        # Vim editor
code file.txt                       # VS Code (if installed)
```

## Makefile Operations

```bash
# Run make targets
make target-name
make -n target-name                 # Dry run (show what would be executed)
make -B target-name                 # Force rebuild
make -j4                            # Parallel execution (4 jobs)
make help                           # Show help (if available)

# Common make targets in this project
make contracts-all                  # Root level
make generate                       # Service level
make test                           # Service level
make build                          # Service level
```

## Monitoring and Debugging

```bash
# Watch command output
watch -n 2 command                  # Run command every 2 seconds
watch -d command                    # Highlight differences

# Tailing logs
tail -f /var/log/app.log
tail -f /var/log/app.log | grep ERROR

# System load
uptime
w                                   # Who is logged in and what they're doing

# Disk I/O
iostat                              # I/O statistics
iotop                               # I/O monitor (Linux, requires sudo)
```
