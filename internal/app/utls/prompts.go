package utls

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ANSI colors
const (
	ColorBold   = "\x1b[1m"
	ColorDim    = "\x1b[2m"
	ColorReset  = "\x1b[22m\x1b[39m"
	ColorCyan   = "\x1b[36m"
	ColorGreen  = "\x1b[32m"
	ColorYellow = "\x1b[33m"
	ColorRed    = "\x1b[31m"
)

const (
	SymbolOk     = "✔"
	SymbolErr    = "✖"
	SymbolWarn   = "▲"
	SymbolInfo   = "•"
	SymbolArrow  = "→"
	SymbolBullet = "·"
)

func Ask(in io.Reader, out io.Writer, isInteractive bool, question string, fallback string) string {
	if !isInteractive {
		if fallback != "" {
			return fallback
		}
		panic(fmt.Sprintf("missing required value for question: %s", question))
	}

	suffix := ""
	if fallback != "" {
		suffix = ColorDim + " (" + fallback + ")" + ColorReset
	}
	fmt.Fprintf(out, "%s?%s %s%s: ", ColorCyan, ColorReset, question, suffix)

	reader := bufio.NewReader(in)
	line, _, err := reader.ReadLine()
	if err != nil {
		if fallback != "" {
			return fallback
		}
		return ""
	}

	ans := strings.TrimSpace(string(line))
	if ans == "" && fallback != "" {
		return fallback
	}
	if ans == "" {
		return Ask(in, out, isInteractive, question, fallback)
	}
	return ans
}

func AskChoice(in io.Reader, out io.Writer, isInteractive bool, question string, choices []string, fallback string) string {
	if !isInteractive {
		if fallback != "" {
			return fallback
		}
		panic(fmt.Sprintf("missing required choice for question: %s", question))
	}

	fmt.Fprintf(out, "%s?%s %s\n", ColorCyan, ColorReset, question)
	for i, c := range choices {
		fmt.Fprintf(out, "  %s%d)%s %s\n", ColorDim, i+1, ColorReset, c)
	}

	fmt.Fprintf(out, "  %sselect 1-%d: %s", ColorDim, len(choices), ColorReset)
	reader := bufio.NewReader(in)
	line, _, err := reader.ReadLine()
	if err != nil {
		if fallback != "" {
			return fallback
		}
		return choices[0]
	}

	ans := strings.TrimSpace(string(line))
	if ans == "" && fallback != "" {
		return fallback
	}
	idx, err := strconv.Atoi(ans)
	if err == nil && idx >= 1 && idx <= len(choices) {
		return choices[idx-1]
	}

	if fallback != "" {
		return fallback
	}

	return AskChoice(in, out, isInteractive, question, choices, fallback)
}

func Confirm(in io.Reader, out io.Writer, isInteractive bool, question string, defaultYes bool) bool {
	if !isInteractive {
		return defaultYes
	}

	hint := "y/N"
	if defaultYes {
		hint = "Y/n"
	}
	fmt.Fprintf(out, "%s?%s %s %s(%s)%s: ", ColorCyan, ColorReset, question, ColorDim, hint, ColorReset)

	reader := bufio.NewReader(in)
	line, _, err := reader.ReadLine()
	if err != nil {
		return defaultYes
	}

	ans := strings.ToLower(strings.TrimSpace(string(line)))
	if ans == "" {
		return defaultYes
	}
	return ans == "y" || ans == "yes"
}
