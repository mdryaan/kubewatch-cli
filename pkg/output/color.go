package output

import (
	"github.com/fatih/color"
)

var (
	Green  = color.New(color.FgGreen, color.Bold)
	Yellow = color.New(color.FgYellow, color.Bold)
	Red    = color.New(color.FgRed, color.Bold)
	Cyan   = color.New(color.FgCyan, color.Bold)
	White  = color.New(color.FgWhite, color.Bold)
	Dim    = color.New(color.Faint)
)

func Healthy(s string) string {
	return Green.Sprint(s)
}

func Warning(s string) string {
	return Yellow.Sprint(s)
}

func Critical(s string) string {
	return Red.Sprint(s)
}

func Info(s string) string {
	return Cyan.Sprint(s)
}

func Bold(s string) string {
	return White.Sprint(s)
}

func Faint(s string) string {
	return Dim.Sprint(s)
}

func DisableColor() {
	color.NoColor = true
}

func StatusColor(status string) string {
	switch status {
	case "Healthy", "Running", "Ready", "True", "Active":
		return Healthy(status)
	case "Warning", "Pending", "Unknown", "Progressing":
		return Warning(status)
	case "Critical", "Failed", "CrashLoopBackOff", "OOMKilled", "Error", "Unhealthy":
		return Critical(status)
	default:
		return status
	}
}
