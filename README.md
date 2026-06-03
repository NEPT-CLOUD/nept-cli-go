# Scalable Go CLI Boilerplate

A production-grade, highly scalable Go Command-Line Interface (CLI) boilerplate built using standard-setting libraries: **Cobra** (CLI parser) and **Viper** (Configuration management).

## Key Features

1. **Modular Subcommand Architecture**: Commands are structured via isolated builder functions, avoiding package-level globals and enabling unit testing.
2. **Unified Configuration**: Config is loaded hierarchically from default values, config files (YAML), and environment variable overrides (prefixed with `NEPT_`).
3. **Structured Logging**: Uses Go's standard library `log/slog` to support debug levels and switches outputs dynamically between human-friendly console output and JSON formatting (for script integration).
4. **Linker Flag Versioning**: Builds inject dynamic git commit hash, date, and version metadata directly during build/compilation time.
5. **Robust Testing Patterns**: Includes patterns to unit test CLI commands by overriding input/output buffer streams and mocking configurations.

---

## Directory Structure

```text
nept-cli-go/
├── cmd/                  # Command line interfaces (Cobra commands)
│   ├── root.go           # Root command & flag parsers
│   ├── hello.go          # Demo subcommand
│   ├── hello_test.go     # Subcommand unit tests
│   └── version.go        # Build version subcommand
├── internal/             # Internal packages (private libraries)
│   ├── app/              # Injectable App container (holds config, logger, io streams)
│   ├── config/           # Configuration parsing & defaults (Viper)
│   └── logger/           # slog configuration & handlers
├── go.mod                # Go module file
├── Makefile              # Build automation
└── README.md             # This document
```

---

## Getting Started

### Prerequisites

- Go `1.25` or higher

### Build

Compile the application to generate the `nept` executable:
```bash
make build
```

This compiles the binary and automatically stamps compile-time information (e.g., current Git commit and timestamp).

### Run

Run the generated binary:
```bash
./nept hello
# Output: Hello, world!
```

#### Override with Flags:
```bash
./nept hello --name Alice --uppercase
# Output: HELLO, ALICE!
```

#### Output JSON format:
```bash
./nept hello --name Bob --format json
# Output:
# {
#   "greeting": "Hello, Bob!",
#   "name": "Bob"
# }
```

### Running Tests

Run all unit tests with race detection and test coverage reporting:
```bash
make test
```

---

## Configuration

Configurations are bound to the `NEPT` environment prefix and are read from the YAML files searched in the following order:
1. File path provided via `--config /path/to/config.yaml`
2. `./.nept.yaml` (Current Working Directory)
3. `~/.nept.yaml` (User's Home Directory)

### Default Config Structure

The default configuration settings map to:
```yaml
environment: production
api_key: your_api_key_here
verbose: false
format: text
```

You can override any variable using environment variables, for example:
```bash
export NEPT_API_KEY="my_secret_token"
export NEPT_VERBOSE="true"
export NEPT_FORMAT="json"
```

---

## How to Add a New Command

1. Create a new file in `cmd/yourcommand.go`.
2. Implement a constructor function returning `*cobra.Command` that takes the `appContainer *app.App`:
   ```go
   package cmd

   import (
       "fmt"
       "github.com/spf13/cobra"
       "github.com/NEPT-CLOUD/nept-cli-go/internal/app"
   )

   func NewYourCommandCmd(appContainer *app.App) *cobra.Command {
       return &cobra.Command{
           Use:   "yourcommand",
           Short: "Does something useful",
           RunE: func(cmd *cobra.Command, args []string) error {
               // Use appContainer.Config, appContainer.Logger, etc.
               appContainer.Logger.Info("Running yourcommand")
               fmt.Fprintln(appContainer.Out, "Success!")
               return nil
           },
       }
   }
   ```
3. Register the subcommand within `cmd/root.go`:
   ```go
   rootCmd.AddCommand(NewYourCommandCmd(appContainer))
   ```
4. Write a unit test using the standard io redirects in `cmd/yourcommand_test.go`.
