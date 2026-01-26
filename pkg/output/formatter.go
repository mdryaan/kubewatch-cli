package output

import (
	"fmt"
	"os"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatPlain Format = "plain"
)

func ParseFormat(s string) Format {
	switch s {
	case "json":
		return FormatJSON
	case "plain":
		return FormatPlain
	default:
		return FormatTable
	}
}

func PrintHeader(title string) {
	fmt.Fprintf(os.Stdout, "\n%s\n", Bold(title))
	fmt.Fprintf(os.Stdout, "%s\n\n", Faint(repeatChar("─", len(title))))
}

func PrintError(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", Critical("✗"), msg)
}

func PrintSuccess(msg string) {
	fmt.Fprintf(os.Stdout, "%s %s\n", Healthy("✓"), msg)
}

func PrintWarning(msg string) {
	fmt.Fprintf(os.Stdout, "%s %s\n", Warning("⚠"), msg)
}

func PrintInfo(msg string) {
	fmt.Fprintf(os.Stdout, "%s %s\n", Info("ℹ"), msg)
}

func repeatChar(c string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += c
	}
	return result
}
