package app

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/NEPT-CLOUD/nept-cli-go/internal/config"
)

// App holds application-wide dependencies. By injecting this container into
// our commands, we avoid global state and make commands completely unit-testable.
type App struct {
	Config *config.Config
	Logger *slog.Logger
	Out    io.Writer
	ErrOut io.Writer
}

// New creates a new instance of the App container.
func New(cfg *config.Config, logger *slog.Logger, out, errOut io.Writer) *App {
	return &App{
		Config: cfg,
		Logger: logger,
		Out:    out,
		ErrOut: errOut,
	}
}

// PrintResult prints the result either as formatted JSON or plain text, depending on configuration.
func (a *App) PrintResult(textValue string, jsonValue interface{}) error {
	if a.Config != nil && a.Config.Format == "json" {
		encoder := json.NewEncoder(a.Out)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(jsonValue); err != nil {
			return fmt.Errorf("failed to encode JSON response: %w", err)
		}
		return nil
	}
	_, err := fmt.Fprintln(a.Out, strings.TrimRight(textValue, "\n"))
	return err
}

// PrintErr prints a structured error in JSON format if requested, otherwise prints plain text to ErrOut.
func (a *App) PrintErr(code string, err error) {
	if err == nil {
		return
	}

	isJSON := false
	if a.Config != nil {
		isJSON = a.Config.Format == "json"
	} else {
		// Fallback check if config is not loaded yet (e.g. command parsing error)
		for i, arg := range os.Args {
			if (arg == "--format" || arg == "-f") && i+1 < len(os.Args) && os.Args[i+1] == "json" {
				isJSON = true
				break
			}
			if arg == "--format=json" || arg == "-f=json" {
				isJSON = true
				break
			}
		}
	}

	if isJSON {
		type ErrorDetail struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		type ErrorResponse struct {
			Status string      `json:"status"`
			Error  ErrorDetail `json:"error"`
		}
		resp := ErrorResponse{
			Status: "error",
			Error: ErrorDetail{
				Code:    code,
				Message: err.Error(),
			},
		}
		encoder := json.NewEncoder(a.ErrOut)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(resp)
	} else {
		fmt.Fprintf(a.ErrOut, "Error: %v\n", err)
	}
}

// AppError represents an application error with a specific error code.
type AppError struct {
	Code string
	Err  error
}

func (e *AppError) Error() string {
	return e.Err.Error()
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError constructs a new AppError.
func NewAppError(code string, err error) *AppError {
	return &AppError{Code: code, Err: err}
}

