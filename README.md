# Nept Go CLI — Complete Documentation

A production-grade, highly scalable Go Command-Line Interface (CLI) designed to connect securely to [nept.cloud](https://nept.cloud) via the `nept-engine-v2-bun` Elysia backend engine. It automates packaging, framework detection, log streaming, custom domain mapping, and database deployments.

---

## Table of Contents
1. [Installation & Uninstallation](#1-installation--uninstallation)
2. [Core Features & Modes](#2-core-features--modes)
3. [Configuration & Authentication Hierarchy](#3-configuration--authentication-hierarchy)
4. [Subcommand Reference & Examples](#4-subcommand-reference--examples)
    - [`nept status`](#nept-status)
    - [`nept login`](#nept-login)
    - [`nept logout`](#nept-logout)
    - [`nept config`](#nept-config)
    - [`nept deploy`](#nept-deploy)
    - [`nept logs` & `nept app logs`](#nept-logs--nept-app-logs)
    - [`nept db deploy`](#nept-db-deploy)
    - [`nept restart` / `delete`](#nept-restart--delete)
    - [`nept domain add`](#nept-domain-add)
5. [Framework Auto-Detection Presets](#5-framework-auto-detection-presets)
6. [In-Memory Zipping & Ignore Rules](#6-in-memory-zipping--ignore-rules)
7. [Cycle-Free Clean Architecture](#7-cycle-free-clean-architecture)
8. [Testing & Subcommand Extension Guide](#8-testing--subcommand-extension-guide)

---

## 1. Installation & Uninstallation

Install the Nept CLI on your system using the official installation scripts:

### macOS / Linux (Shell)

You can download and run the installer shell script via `curl` or `wget`:

```bash
curl -fsSL https://raw.githubusercontent.com/NEPT-CLOUD/nept-cli-go/main/install.sh | sh
```

### Windows (PowerShell)

You can run the installer PowerShell script directly:

```powershell
powershell -c "irm https://raw.githubusercontent.com/NEPT-CLOUD/nept-cli-go/main/install.ps1 | iex"
```

### Upgrading

To upgrade the Nept CLI to the latest version, run the appropriate script for your OS:

#### macOS / Linux (Shell)
```bash
curl -fsSL https://raw.githubusercontent.com/NEPT-CLOUD/nept-cli-go/main/upgrade.sh | sh
```

#### Windows (PowerShell)
```powershell
powershell -c "irm https://raw.githubusercontent.com/NEPT-CLOUD/nept-cli-go/main/upgrade.ps1 | iex"
```

You can also run the local script directly or pass `-f`/`--force` to force reinstalling the latest version:
* macOS/Linux: `./upgrade.sh` (or `./upgrade.sh --force`)
* Windows: `.\upgrade.ps1` (or `.\upgrade.ps1 -Force`)

### Uninstallation

If you need to completely remove the Nept CLI, its configurations, and the downloaded skill files from your system, run the appropriate command for your OS:

#### macOS / Linux (Shell)
```bash
curl -fsSL https://raw.githubusercontent.com/NEPT-CLOUD/nept-cli-go/main/uninstall.sh | sh
```

#### Windows (PowerShell)
```powershell
powershell -c "irm https://raw.githubusercontent.com/NEPT-CLOUD/nept-cli-go/main/uninstall.ps1 | iex"
```

#### What gets cleaned up:
* The `nept` executable binary from your system PATH (e.g. `/usr/local/bin/nept` or `~/.nept/bin/nept.exe`).
* The local skill guidelines folder (e.g. `~/.nept/skill` or `~/.nept` directory).
* Optional prompt to remove your global user configuration file (`~/.nept.yaml`).

---

## 2. Core Features & Modes

The CLI dynamically detects its execution environment and switches between two operational modes:

| Mode | Trigger | Behavior |
| --- | --- | --- |
| **Human Mode** | Interactive Terminal (TTY) | Vibrant ANSI colors, friendly log prefixes, interactive prompts, and live SSE build log streaming. |
| **Agent Mode** | `--format json` / `-f json`, `JSON=true`, or `OUTPUT_FORMAT=json` | Raw JSON output to `stdout`, **no interactive blocks or prompts**, suppressed colors. Failures write standard `{"error": "...", "code": 1}` to `stderr` and exit non-zero. |

This makes the CLI highly interactive for developers during local workflows, and fully automated for pipelines (CI/CD) and AI agents.

---

## 3. Configuration & Authentication Hierarchy

Authentication and connection parameters are resolved hierarchically from three distinct layers, in order of precedence:

1. **Environment Variables**:
   - `NEPT_API_KEY`: Overrides active API Key.
   - `NEPT_USER_ID`: Overrides kubernetes namespace (UUID).
   - `NEPT_API_URL`: Overrides target engine URL.
2. **Configuration File** (`.nept.yaml` in working directory, or `~/.nept.yaml` in home directory):
   - Managed via Viper. Values are loaded dynamically.
3. **OS Keyring** (via `keyring` integration):
   - Credentials saved securely via `nept login` are loaded as fallback if not present in config/env.

---

## 4. Subcommand Reference & Examples

### `nept status`
Checks the reachability and health of the backend engine (`/api/health`).

#### Human Mode Example:
```bash
$ nept status
✔ Engine online  http://localhost:8000
  status    ok
  uptime    4827s
  time      2026-06-08T15:36:04Z
```

#### Agent Mode Example:
```bash
$ nept status --format json
{
  "connected": true,
  "status": "ok",
  "uptime": 4827,
  "timestamp": "2026-06-08T15:36:04Z"
}
```

---

### `nept login`
Authenticates with an API Key, verifies validity with the backend, retrieves the corresponding `userId` (namespace), and registers credentials securely into the OS Keychain.

```bash
# Validates key and caches both key and userId
$ nept login -k nept_8d6468ec36ab3601d1fc7c2c9ba6cf5dbee349bba62c8904217dd208603335cb
Login successful. API key saved to keychain.
```

---

### `nept logout`
Logs out the user by clearing the saved API key and User ID credentials from the OS Keychain.

#### Human Mode Example:
```bash
$ nept logout
Logout successful. Credentials cleared from keychain.
```

#### Agent Mode Example:
```bash
$ nept logout --format json
{
  "status": "success",
  "message": "Credentials cleared from keychain."
}
```

---

### `nept config`
Reads, sets, or lists Viper configuration variables. Resolves sources dynamically.

#### Config List:
```bash
$ nept config list
Effective configuration
  api_url   http://localhost:8000 (default)
  user_id   811f9ad3-21a2-4d8d-8fa9-4d6542e7fead (keyring)
  api_key   nept_8d6...35cb (keyring)
  file      /Users/nazmussamir/.nept.yaml
```

#### Config Set:
```bash
$ nept config set api_url https://server.nept.cloud
✔ Set api_url = https://server.nept.cloud
```

#### Config Get:
```bash
$ nept config get user_id
811f9ad3-21a2-4d8d-8fa9-4d6542e7fead
```

---

### `nept deploy`
Packages a directory in-memory, auto-detects the framework, resolves git branch/commit context, uploads a base64-encoded zip payload to `/api/deploy`, and live-streams Server-Sent Event (SSE) build logs to completion.

#### Interactive Human Mode:
```bash
$ nept deploy ./my-next-app
? Detected framework: Next.js. Use it? (Y/n): y
? Project name (my-next-app): my-next-app
Packaging project...
✔ Packaged 45 files (1.2 MB)
Uploading & starting build...
✔ Build started

  project      my-next-app
  deployment   dep-f37b92ac
  logs         logs-9a8b7c

Build logs:
  · [2026-06-08T15:45:01Z] npm install completed
  · [2026-06-08T15:45:10Z] Next.js build completed
  · [2026-06-08T15:45:12Z] Image built & pushed to registry
✔ Deployed
  → https://my-next-app.nept.cloud
```

#### Silent Agent Mode (useful for CI/CD):
```bash
$ nept deploy ./my-next-app --yes --format json
{
  "success": true,
  "message": "Build started",
  "logsID": "logs-9a8b7c",
  "timestamp": "2026-06-08T15:45:00Z",
  "projectId": "proj-9a8b7c",
  "deploymentId": "dep-f37b92ac",
  "domain": "my-next-app.nept.cloud"
}
```

---

### `nept logs` & `nept app logs`
- `nept logs <logsId>`: Connects to SSE log stream `/api/logs/<logsId>` and outputs live build logs. Exits non-zero if build fails.
- `nept app logs <deploymentId>`: Fetches historical runtime logs for deployed containers.

```bash
# Stream build logs
$ nept logs logs-9a8b7c
Streaming logs for logs-9a8b7c
  · npm install completed
  · Next.js build completed
✔ Stream ended
```

```bash
# Fetch container runtime logs
$ nept app logs dep-f37b92ac
[2026-06-08T15:46:00Z] Server listening on port 3000
[2026-06-08T15:46:12Z] GET /api/health - 200 OK
```

---

### `nept db deploy`
Deploys isolated databases (`postgres`, `mysql`, `mongodb`, `redis`) to the cluster, configuring storage volumes, CPU limits, and memory limits.

```bash
$ nept db deploy --type postgres --name my-db --volume 20 --cpu 2 --memory 1024
Deploying postgres database...
✔ Database 'my-db' deployed

  id          db-811f9ad3
  host        my-db.nept.cloud
  port        5432
  username    admin
  password    p@ssw0rd123!
  url         postgresql://admin:p%40ssw0rd123%21@my-db.nept.cloud:5432/my-db

▲ Store the password now — it cannot be retrieved later.
```

---

### `nept restart` / `delete`
Manages application lifecycle.

```bash
# Gracefully restarts a running container deployment
$ nept restart my-next-app
✔ Deployment restarted for my-next-app
```

```bash
# Deletes deployment and cleans up namespace resources
$ nept delete my-next-app
? Delete my-next-app and all its resources? (y/N): y
Deleting my-next-app...
✔ Deleted all for my-next-app
```

---

### `nept domain add`
Attaches custom domains to deployed projects and prints the ownership token and challenge validation TXT records.

```bash
$ nept domain add proj-9a8b7c example.com
Attaching example.com to project proj-9a8b7c...
✔ Domain example.com registered

Add these DNS records:
  CNAME  example.com                 →  origin.nept.cloud
  TXT    _cf-custom-hostname.example.com   challenges-token-val
  TXT    _acme-challenge.example.com       ssl-verification-val

SSL: pending_validation (txt)
```

---

## 5. Framework Auto-Detection Presets

The CLI automatically inspects directory files to match indicator dependencies against default presets:

| Indicator File | Match Criteria | Framework Preset | Build Command | Run Command | Default Port |
| --- | --- | --- | --- | --- | --- |
| `package.json` | dependency: `next` | `Next.js` | `npm install`, `npm run build` | `npm start` | 3000 |
| `package.json` | dependency: `react` | `React` | `npm install`, `npm run build` | `npm start` | 3000 |
| `package.json` | dependency: `express` | `Express.js` | `npm install` | `npm start` | 3000 |
| `requirements.txt` | content: `fastapi` | `FastAPI` | `pip install -r requirements.txt` | `uvicorn main:app ...` | 8000 |
| `go.mod` | present | `Go` | `go build -o main .` | `./main` | 8080 |
| `Cargo.toml` | present | `Rust` | `cargo build --release` | `./app` | 8080 |
| `index.html` | present | `html-css-js` | None | `serve -s . -l 80` | 80 |

---

## 6. In-Memory Zipping & Ignore Rules

When packaging directories, the CLI compresses files in-memory using `archive/zip` to optimize bandwidth and speed.

### Ignored Assets
By default, the packaging engine filters out typical development clutter:
```text
node_modules, .git, .next, .output, dist, build, target, .env, .DS_Store, *.log
```

### Custom Ignore Files
In addition, the packager loads patterns sequentially from:
1. `.gitignore`
2. `.neptignore`

These patterns are parsed, anchored or unanchored, and translated into compiled regular expressions. For instance, `/dist` anchors matching to the root directory only, while `*.log` matches files recursively.

---

## 7. Cycle-Free Clean Architecture

To avoid Go package import loops (`internal/app` importing `internal/config`, and `internal/app/utils` importing `internal/app` to call `App`), the utility package defines a decoupled `APIContainer` interface:

```go
// Located in internal/app/utils/api.go
type APIContainer interface {
	GetAPIURL() string
	ResolveAPIKey() (string, error)
	GetStdout() io.Writer
}
```

The unified `App` structure (in [internal/app/app.go](file:///Users/nazmussamir/Documents/nept-v2/nept-cli/nept-cli-go/internal/app/app.go)) implements this interface:

```go
func (a *App) GetAPIURL() string {
	if a.Config != nil {
		return a.Config.APIURL
	}
	return ""
}

func (a *App) GetStdout() io.Writer {
	return a.Out
}
```

This ensures `internal/app/utils` has zero compile-time dependencies on `internal/app`, maintaining a clean, directional dependency tree.

---

## 8. Testing & Subcommand Extension Guide

All subcommands are completely unit-testable by injecting a mocked `app.App` container containing buffered IO streams.

### Example: Testing a Custom Subcommand
Create a test file alongside your command (e.g. `cmd/hello_test.go`). Configure custom buffers to assert console prints:

```go
package cmd

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/NEPT-CLOUD/nept-cli-go/internal/app"
	"github.com/NEPT-CLOUD/nept-cli-go/internal/config"
)

func TestHelloCmd(t *testing.T) {
	t.Run("greeting with flag and uppercase", func(t *testing.T) {
		outBuf := new(bytes.Buffer)
		errBuf := new(bytes.Buffer)

		appContainer := &app.App{
			Config: &config.Config{
				Environment: "test",
				Format:      "text",
				Verbose:     false,
			},
			Logger: slog.New(slog.NewTextHandler(errBuf, nil)),
			Out:    outBuf,
			ErrOut: errBuf,
		}

		cmd := NewHelloCmd(appContainer)
		cmd.SetOut(outBuf)
		cmd.SetErr(errBuf)
		cmd.SetArgs([]string{"--name", "Alice", "--uppercase"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected execution error: %v", err)
		}

		got := strings.TrimSpace(outBuf.String())
		expected := "HELLO, ALICE!"
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})
}
```

### Running Tests
Execute the Go test suite to run all assertions across commands and packaging engines:
```bash
$ go test -v ./...
```
