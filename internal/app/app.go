package app

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/NEPT-CLOUD/nept-cli-go/internal/config"
	"github.com/zalando/go-keyring"
)

// App holds application-wide dependencies. By injecting this container into
// our commands, we avoid global state and make commands completely unit-testable.
type App struct {
	Config *config.Config
	Logger *slog.Logger
	Out    io.Writer
	ErrOut io.Writer
	In     io.Reader
}

// New creates a new instance of the App container.
func New(cfg *config.Config, logger *slog.Logger, out, errOut io.Writer, in io.Reader) *App {
	return &App{
		Config: cfg,
		Logger: logger,
		Out:    out,
		ErrOut: errOut,
		In:     in,
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

// GetAPIURL returns the configured API URL.
func (a *App) GetAPIURL() string {
	if a.Config != nil {
		return a.Config.APIURL
	}
	return ""
}

// GetStdout returns the output stream.
func (a *App) GetStdout() io.Writer {
	return a.Out
}


// ResolveAPIKey resolves the active API Key from config, environment, or keyring.
func (a *App) ResolveAPIKey() (string, error) {
	if a.Config != nil && a.Config.APIKey != "" {
		return a.Config.APIKey, nil
	}
	key, err := keyring.Get("nept", "api-key")
	if err == nil && key != "" {
		return key, nil
	}
	return "", fmt.Errorf("not logged in. Run 'nept login -k <api_key>' or set NEPT_API_KEY environment variable")
}

// ResolveUserID resolves the active User ID from config, environment, keyring, or by validating the key.
func (a *App) ResolveUserID() (string, error) {
	if a.Config != nil && a.Config.UserID != "" {
		return a.Config.UserID, nil
	}
	uid, err := keyring.Get("nept", "user-id")
	if err == nil && uid != "" {
		return uid, nil
	}

	apiKey, err := a.ResolveAPIKey()
	if err != nil {
		return "", err
	}

	a.Logger.Debug("UserID not configured; fetching from validate API")
	resolvedUid, err := a.FetchUserIDFromAPI(apiKey)
	if err != nil {
		return "", fmt.Errorf("failed to automatically resolve user_id: %w", err)
	}

	_ = keyring.Set("nept", "user-id", resolvedUid)
	return resolvedUid, nil
}

// FetchUserIDFromAPI queries the validation endpoint to retrieve the userId associated with the key.
func (a *App) FetchUserIDFromAPI(apiKey string) (string, error) {
	apiURL := "https://server.nept.cloud"
	if a.Config != nil && a.Config.APIURL != "" {
		apiURL = a.Config.APIURL
	}
	req, err := http.NewRequest("GET", apiURL+"/api/keys/validate", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API validation failed with status %d", resp.StatusCode)
	}
	var resBody struct {
		UserID string `json:"userId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&resBody); err != nil {
		return "", err
	}
	if resBody.UserID == "" {
		return "", fmt.Errorf("no userId returned from API")
	}
	return resBody.UserID, nil
}


