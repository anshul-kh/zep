# Zep

**Zep** is a lightweight process manager written in Go.

It is designed to easily run and manage background processes, services, cron jobs, shell commands, and APIs with minimal setup.

> ⚠️ **Note:** Zep is currently supported **only on Unix-based systems** (e.g. Linux, macOS). Windows support is not available.


---

## ✨ Features

- 🧠 Simple CLI interface
- 🚀 Run and manage background services or shell commands
- 🧩 Support for naming and watching processes
- 📊 View process stats
- 🛑 Stop processes by ID or name
- 🐚 Shell autocompletion support

---

## 📦 Installation

Follow these steps to install Zep from source.

### 1. Clone the Repository

```bash
git clone https://github.com/anshul-kh/zep.git
cd zep
```

---

### 2. Install Dependencies

Run the dependency installation script:

```bash
bash ./scripts/install-deps.sh
```

---

### 3. Build and Install

Use the provided Makefile to build the CLI and install the system service:

```bash
make install-service build-cli
```

This will:

- Install the Zep systemd service (or equivalent)
- Build the `zep` CLI binary

---

## ⚙️ Usage

```
zep [command] [flags]
```

### 📋 Available Commands

| Command      | Description                                        |
| ------------ | -------------------------------------------------- |
| `completion` | Generate autocompletion script for your shell      |
| `help`       | Help about any command                             |
| `list`       | List all running or registered processes           |
| `run`        | Run and manage a binary using the binary path      |
| `stats`      | Show stats for a process by ID or `--name`         |
| `stop`       | Stop a running process by ID or `--name`           |
| `watch`      | Watch live output from a process by ID or `--name` |

### 🌐 Global Flags

- `-h`, `--help` — Show help for Zep.
- `--name string` — Assign or filter by name for a process.

---

## 🚀 Examples

### Run a Process

```bash
zep run my-binary --name my-service
```

### Run a Shell Command

```bash
zep run "ls -l" --name my-service --shell-cmd
```

### List Running Processes

```bash
zep list
```

### Get Process Stats

```bash
zep stats --name my-service
```

### Stop a Process

```bash
zep stop --name my-service
```

### Watch Process Output

```bash
zep watch --name my-service
```

---

## 🧪 Test / Demo

### Build Test Binary

```bash
make build-test
```

### Run Test Binary

```bash
zep run ./bin/test --name my-service
```

---

## 🛠️ Requirements

- Go 1.18+
- `make` installed
- `systemd` (for service installation on Linux)

---

> **Zep** makes managing background processes fast, simple, and reliable.
