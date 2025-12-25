# System Commands Reference

This project is developed on **Linux** (also supports macOS). Standard Unix commands are available.

## File System Operations

```bash
# List files
ls -la

# List directory tree
tree -L 3

# Change directory
cd path/to/directory

# Create directory
mkdir -p path/to/directory

# Remove files/directories
rm file.txt
rm -rf directory/

# Copy files
cp source.txt destination.txt
cp -r source_dir/ dest_dir/

# Move/rename files
mv old_name.txt new_name.txt

# Find files
find . -name "*.go"
find . -type f -name "*.proto"

# Check if file exists
test -f file.txt && echo "exists"
```

## Text Processing

```bash
# Search in files
grep -r "pattern" .
grep -n "pattern" file.txt

# View file contents
cat file.txt
less file.txt
head -n 20 file.txt
tail -n 20 file.txt

# Count lines
wc -l file.txt

# Text manipulation
sed 's/old/new/g' file.txt
awk '{print $1}' file.txt
```

## Process Management

```bash
# List processes
ps aux
ps aux | grep service-name

# Kill process
kill PID
kill -9 PID

# Background/foreground jobs
command &          # Run in background
jobs              # List background jobs
fg %1             # Bring job to foreground
```

## Network Operations

```bash
# Check port usage
netstat -tlnp | grep :8000
lsof -i :8000

# HTTP requests
curl http://localhost:8000/health
curl -X POST http://localhost:8000/v1/symbols -d '{"name":"test"}'

# DNS lookup
nslookup domain.com
dig domain.com
```

## Git Operations

```bash
# Status and history
git status
git log --oneline
git diff

# Branching
git branch
git checkout -b feature-branch
git checkout main

# Staging and committing
git add .
git add file.txt
git commit -m "message"

# Remote operations
git push origin branch-name
git pull origin main
git fetch

# Stashing
git stash
git stash pop
```

## Docker Operations

```bash
# List containers
docker ps
docker ps -a

# View logs
docker logs container-name
docker logs -f container-name  # Follow logs

# Docker Compose
docker-compose up
docker-compose up -d           # Detached mode
docker-compose down
docker-compose logs service-name
docker-compose ps

# Execute in container
docker exec -it container-name bash
docker exec container-name command
```

## Go Commands

```bash
# Module operations
go mod tidy
go mod download
go mod verify

# Building
go build ./...
go build -o output_name ./cmd/service

# Testing
go test ./...
go test -v ./internal/biz/
go test -run TestName
go test -race ./...
go test -cover ./...

# Running
go run ./cmd/service/

# Formatting
gofmt -s -w .
go fmt ./...

# Vetting
go vet ./...

# List packages
go list ./...
go list -m all
```

## Environment Variables

```bash
# View environment
env
printenv

# Set variable
export VAR_NAME=value

# Unset variable
unset VAR_NAME

# Check Go environment
go env
go env GOPATH
go env GOROOT
```

## Permissions

```bash
# Change file permissions
chmod +x script.sh
chmod 644 file.txt

# Change ownership
chown user:group file.txt

# View permissions
ls -l file.txt
```

## Disk Usage

```bash
# Check disk space
df -h

# Check directory size
du -sh directory/
du -h --max-depth=1
```

## System Information

```bash
# OS information
uname -a
cat /etc/os-release

# CPU information
lscpu
cat /proc/cpuinfo

# Memory information
free -h
cat /proc/meminfo
```

## Useful Shortcuts

```bash
# Command history
history
!123          # Run command number 123
!!            # Run last command
!$            # Last argument of previous command

# Pipes and redirection
command1 | command2
command > file.txt
command >> file.txt  # Append
command 2>&1        # Redirect stderr to stdout
```
