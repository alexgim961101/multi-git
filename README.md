# Multi-Git

A CLI tool for efficiently managing multiple Git repositories. Helps DevOps engineers automate repetitive tasks across multiple repositories.

## 📋 Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Usage](#usage)
- [Examples](#examples)
- [Contributing](#contributing)
- [License](#license)

<a id="features"></a>
## ✨ Features

- **Batch Repository Cloning**: Clone multiple Git repositories at once
- **Batch Branch Checkout**: Checkout the same branch across all managed repositories simultaneously
- **Tag Management**: Create and push tags to specific branches across multiple repositories simultaneously
- **Force Push**: Support for force push to resolve branch conflicts during release deployment
- **Command Execution**: Execute the same shell commands/scripts across all repositories

<a id="installation"></a>
## 🚀 Installation

### Requirements

- Go 1.24 or higher
- Git 2.0 or higher

### Build from Source

```bash
# Clone the repository
git clone https://github.com/lotto/multi-git.git
cd multi-git

# Build
go build -o multi-git cmd/multi-git/main.go

# Install (optional)
sudo mv multi-git /usr/local/bin/
```

### Installation Script (Recommended)

The easiest way to install multi-git is using the installation script:

```bash
# Clone the repository
git clone https://github.com/lotto/multi-git.git
cd multi-git

# Install to /usr/local/bin (requires sudo)
./scripts/install.sh

# Or install to user directory (no sudo required)
./scripts/install.sh --user

# Or install to custom path
./scripts/install.sh --prefix=/opt/bin
```

**Installation Options:**
- Default (`./scripts/install.sh`): Installs to `/usr/local/bin` (requires sudo)
- `--user`: Installs to `~/.local/bin` (no sudo required)
- `--prefix=PATH`: Installs to custom path

**Note:** If you use `--user` option, make sure `~/.local/bin` is in your PATH. Add this to your `~/.bashrc` or `~/.zshrc`:

```bash
export PATH="$PATH:$HOME/.local/bin"
```

**Verify Installation:**

```bash
multi-git --version
```

### Uninstall

To uninstall multi-git:

```bash
# Uninstall from default location (/usr/local/bin)
./scripts/uninstall.sh

# Uninstall from user directory (~/.local/bin)
./scripts/uninstall.sh --user

# Uninstall and remove configuration files
./scripts/uninstall.sh --all
```

**Uninstall Options:**
- Default (`./scripts/uninstall.sh`): Removes binary from `/usr/local/bin`
- `--user`: Removes binary from `~/.local/bin`
- `--all`: Also removes configuration files from `~/.multi-git/`
- `--prefix=PATH`: Removes binary from custom path

### Binary Download (Coming Soon)

You can download binaries for your operating system from the releases page.

<a id="quick-start"></a>
## 🏃 Quick Start

1. **Create Configuration File**

```bash
mkdir -p ~/.multi-git
cat > ~/.multi-git/config.yaml << EOF
config:
  base_dir: ~/repositories
  default_remote: origin
  parallel_workers: 3

repositories:
  - name: backend-service
    url: https://github.com/org/backend-service.git
  
  - name: frontend-app
    url: https://github.com/org/frontend-app.git
EOF
```

2. **Clone Repositories**

```bash
multi-git clone
```

3. **Checkout Branch**

```bash
multi-git checkout release/v1.0.0
```

<a id="configuration"></a>
## ⚙️ Configuration

The configuration file is located at `~/.multi-git/config.yaml` by default. You can specify a different path using the `--config` flag.

### Configuration File Structure

```yaml
config:
  base_dir: ~/repositories      # Base directory for cloning repositories
  default_remote: origin         # Default remote name
  parallel_workers: 3            # Number of parallel operations

repositories:
  - name: backend-service        # Repository name
    url: https://github.com/org/backend-service.git  # Repository URL
    path: backend                # Optional path override
  
  - name: frontend-app
    url: https://github.com/org/frontend-app.git
    # If path is not specified, name is used
```

### Repository URL Formats

- HTTPS: `https://github.com/org/repo.git`
- SSH: `git@github.com:org/repo.git`

<a id="usage"></a>
## 📖 Usage

### `clone` - Clone Repositories

Clone multiple repositories at once.

```bash
multi-git clone [flags]
```

**Flags:**
- `--config, -c`: Configuration file path (default: `~/.multi-git/config.yaml`)
- `--skip-existing`: Skip repositories that already exist (default: `true`)
- `--parallel, -p`: Number of parallel clones (default: `3`)
- `--depth`: Shallow clone depth (optional)

**Examples:**
```bash
# Basic clone
multi-git clone

# Specify number of parallel clones
multi-git clone --parallel 5

# Re-clone existing repositories
multi-git clone --skip-existing=false
```

### `checkout` - Batch Branch Checkout

Checkout the same branch across all managed repositories at once.

```bash
multi-git checkout <branch-name> [flags]
```

**Flags:**
- `--create, -c`: Create branch if it doesn't exist
- `--force, -f`: Force checkout, discarding local changes
- `--fetch`: Fetch from remote before checkout

**Examples:**
```bash
# Checkout branch
multi-git checkout release/v1.0.0

# Create branch if it doesn't exist
multi-git checkout feature/new-feature --create

# Fetch before checkout
multi-git checkout release/v1.0.0 --fetch
```

### `tag` - Tag Management

Create and push tags to specific branches across multiple repositories simultaneously.

```bash
multi-git tag --branch <branch> --name <tag-name> [flags]
```

**Flags:**
- `--branch, -b`: Branch name to create tag on (required for creation, optional for deletion)
- `--name, -n`: Tag name (required)
- `--message, -m`: Tag message
- `--push, -p`: Push tag to remote
- `--force, -f`: Overwrite existing tag
- `--delete, -d`: Delete tag

**Examples:**
```bash
# Create a tag
multi-git tag --branch release/v1.0.0 --name v1.0.0

# Create and push tag
multi-git tag --branch release/v1.0.0 --name v1.0.0 --push --message "Release v1.0.0"

# Delete a tag
multi-git tag --name v1.0.0 --delete --push
```

### `push` - Force Push

Perform force push on specific branches across multiple repositories.

```bash
multi-git push --branch <branch> --force [flags]
```

**Flags:**
- `--branch, -b`: Branch name to push (required, supports `local:remote` format)
- `--force, -f`: Force push (required)
- `--remote, -r`: Remote name (default: `origin`)
- `--dry-run`: Simulate without actually pushing
- `--yes, -y`: Skip confirmation prompt

**Examples:**
```bash
# Force push (with confirmation prompt)
multi-git push --branch release/v1.0.0 --force

# Push local branch to different remote branch name
multi-git push --branch master:aging --force

# Skip confirmation prompt
multi-git push --branch release/v1.0.0 --force --yes

# Dry-run mode (simulation only)
multi-git push --branch release/v1.0.0 --force --dry-run
```

### `exec` - Execute Commands

Execute the same shell commands/scripts across all repositories.

```bash
multi-git exec <command> [flags]
```

**Flags:**
- `--parallel, -p`: Number of parallel operations (default: config value, 0=sequential)
- `--fail-fast`: Stop on first failure
- `--shell, -s`: Shell to use (default: `/bin/sh`)
- `--dry-run`: Simulate without actually executing
- `--show-output, -o`: Show command output (default: `true`)

**Examples:**
```bash
# Run npm install in all repositories
multi-git exec "npm install"

# Check git status in all repositories
multi-git exec "git status"

# Create a file in all repositories
multi-git exec "touch .gitkeep"

# Sequential execution (no parallel)
multi-git exec "npm test" --parallel 0

# Stop on first failure
multi-git exec "make build" --fail-fast

# Dry-run mode (no actual execution)
multi-git exec "rm -rf node_modules" --dry-run

# Hide output
multi-git exec "npm install" --show-output=false
```

<a id="examples"></a>
## 💡 Examples

### Scenario 1: Release Preparation

```bash
# 1. Checkout release branch across all repositories
multi-git checkout release/v1.0.0 --fetch

# 2. Create and push release tag
multi-git tag --branch release/v1.0.0 --name v1.0.0 --push --message "Release v1.0.0"
```

### Scenario 2: Resolve Conflicts After Deployment

```bash
# Force push to resolve branch conflicts
multi-git push --branch release/v1.0.0 --force --yes
```

### Scenario 3: New Project Setup

```bash
# 1. Clone all repositories
multi-git clone

# 2. Checkout development branch
multi-git checkout develop --fetch
```

### Scenario 4: Install Dependencies Across All Repositories

```bash
# Run npm install in all repositories
multi-git exec "npm install"

# Or use yarn
multi-git exec "yarn install"
```

### Scenario 5: Build and Test Across All Repositories

```bash
# Build in all repositories
multi-git exec "npm run build"

# Test in all repositories (stop on first failure)
multi-git exec "npm test" --fail-fast
```

### Scenario 6: Create Common Files Across All Repositories

```bash
# Create .gitkeep file in all repositories
multi-git exec "touch .gitkeep"

# Create .env.example file in all repositories
multi-git exec "cp .env .env.example"
```

## 🛠️ Development

### Project Structure

```
multi-git/
├── cmd/
│   └── multi-git/          # CLI entry point
├── internal/
│   ├── commands/           # Command implementations
│   ├── config/             # Configuration management
│   ├── repository/         # Repository management
│   ├── git/                # Git operations
│   └── shell/              # Shell command execution
├── pkg/
│   └── errors/             # Error types
└── docs/                    # Documentation
```

### Build

```bash
go build -o multi-git cmd/multi-git/main.go
```

### Test

```bash
go test ./...
```

<a id="contributing"></a>
## 🤝 Contributing

Contributions are welcome! Please create an issue or submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

<a id="license"></a>
## 📝 License

This project is licensed under the MIT License.

## 📚 Related Documentation

- [PRD](./docs/PRD.md) - Product Requirements Document
- [Tech Spec](./docs/TECH_SPEC.md) - Technical Specification

## 🐛 Bug Reports

Found a bug? Please [create an issue](https://github.com/lotto/multi-git/issues).
