package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type APIContainer interface {
	GetAPIURL() string
	ResolveAPIKey() (string, error)
	GetStdout() io.Writer
}

func CallAPI(appContainer APIContainer, method, path string, reqBody interface{}, respDest interface{}) (int, error) {
	// 1. Resolve URL
	apiURL := appContainer.GetAPIURL()
	if apiURL == "" {
		apiURL = "https://server.nept.cloud"
		if BackendUrl != "" {
			apiURL = BackendUrl
		}
	}

	url := apiURL + path

	// 2. Prepare body
	var bodyReader io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return 0, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// 3. Resolve and set API key
	apiKey, err := appContainer.ResolveAPIKey()
	if err == nil && apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	// 4. Make request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("cannot reach engine at %s (%w)", apiURL, err)
	}
	defer resp.Body.Close()

	// 5. Read response
	var buf bytes.Buffer
	_, _ = ioCopy(&buf, resp.Body)
	respBytes := buf.Bytes()

	if resp.StatusCode >= 400 {
		// Try to parse error message
		var errBody map[string]interface{}
		if err := json.Unmarshal(respBytes, &errBody); err == nil {
			if msg := parseErrorBody(errBody); msg != "" {
				return resp.StatusCode, fmt.Errorf("%s", msg)
			}
		}
		return resp.StatusCode, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBytes))
	}

	if respDest != nil && len(respBytes) > 0 {
		if err := json.Unmarshal(respBytes, respDest); err != nil {
			// Fallback: if respDest is a pointer to a string, assign raw response
			if pStr, ok := respDest.(*string); ok {
				*pStr = string(respBytes)
				return resp.StatusCode, nil
			}
			return resp.StatusCode, fmt.Errorf("failed to decode response: %w (body: %s)", err, string(respBytes))
		}
	}

	return resp.StatusCode, nil
}

func ioCopy(dst io.Writer, src io.Reader) (written int64, err error) {
	buf := make([]byte, 32*1024)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = fmt.Errorf("invalid write")
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = ioErrShortWrite
				break
			}
		}
		if er != nil {
			if er.Error() != "EOF" {
				err = er
			}
			break
		}
	}
	return written, err
}

var ioErrShortWrite = fmt.Errorf("short write")

func parseErrorBody(body map[string]interface{}) string {
	if errVal, ok := body["error"]; ok {
		if errMap, ok := errVal.(map[string]interface{}); ok {
			if msg, ok := errMap["message"].(string); ok {
				return msg
			}
		}
		if errStr, ok := errVal.(string); ok {
			return errStr
		}
	}
	if details, ok := body["details"].(string); ok {
		return details
	}
	if msg, ok := body["message"].(string); ok {
		return msg
	}
	return ""
}

type LogEntry struct {
	Level         string `json:"level"`
	Message       string `json:"message"`
	Status        string `json:"status"`
	Timestamp     string `json:"timestamp"`
	LogsSecretKey string `json:"logs_secrect_key"`
}

func LevelColor(level string) string {
	switch strings.ToLower(level) {
	case "error":
		return ColorRed
	case "warn", "warning":
		return ColorYellow
	case "success":
		return ColorGreen
	case "debug":
		return ColorDim
	default:
		return ""
	}
}

func StreamBuildLogs(appContainer APIContainer, logsID string, jsonMode bool) (bool, []LogEntry, error) {
	apiURL := appContainer.GetAPIURL()
	if apiURL == "" {
		apiURL = "https://server.nept.cloud"
		if BackendUrl != "" {
			apiURL = BackendUrl
		}
	}

	url := apiURL + "/api/logs/" + logsID
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, nil, err
	}

	apiKey, err := appContainer.ResolveAPIKey()
	if err == nil && apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, nil, fmt.Errorf("cannot reach engine at %s (%w)", apiURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, nil, fmt.Errorf("failed to open log stream (HTTP %d)", resp.StatusCode)
	}

	var entries []LogEntry
	failed := false

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return false, nil, err
		}

		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data:") {
			continue
		}

		dataJSON := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if dataJSON == "" {
			continue
		}

		var obj LogEntry
		if err := json.Unmarshal([]byte(dataJSON), &obj); err != nil {
			continue
		}

		// Skip handshake
		if obj.LogsSecretKey != "" && obj.Message == "" {
			continue
		}

		entries = append(entries, obj)
		if obj.Status == "failed" {
			failed = true
		}

		if !jsonMode && obj.Message != "" {
			color := LevelColor(obj.Level)
			fmt.Fprintf(appContainer.GetStdout(), "  %s%s%s %s%s%s\n", ColorDim, SymbolBullet, ColorReset, color, obj.Message, ColorReset)
		}
	}

	return failed, entries, nil
}
