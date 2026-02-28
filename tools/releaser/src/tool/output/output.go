package output

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
	colorCyan   = "\033[36m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
	colorRed    = "\033[91m"
	colorBlue   = "\033[34m"
	colorSteel  = "\033[38;5;110m"
	colorOrange = "\033[38;5;208m"
	colorWhite  = "\033[97m"
	colorGray   = "\033[90m"
)

var enableColor = supportsColor()
var verbosityLevel int

func Blank() {
	fmt.Println()
}

func Ask(prompt string) string {
	Blank()
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	Blank()
	return strings.TrimSpace(answer)
}

func Info(msg string) {
	printLine(os.Stdout, "INFO", colorCyan, msg)
}

func Continue(msg string) {
	fmt.Fprintln(os.Stdout, "      ", msg)
}

func Verbose(msg string) {
	if verbosityLevel < 1 {
		return
	}
	printLine(os.Stdout, "DEBUG", colorBlue, msg)
}

func VerboseList(title string, items []string, max int) {
	if verbosityLevel < 1 || len(items) == 0 {
		return
	}
	if max <= 0 {
		max = 10
	}

	Verbose(fmt.Sprintf("%s (%d)", title, len(items)))
	limit := len(items)
	if limit > max {
		limit = max
	}
	for i := 0; i < limit; i++ {
		Verbose("  - " + items[i])
	}
	if len(items) > max {
		Verbose(fmt.Sprintf("  ... +%d more", len(items)-max))
	}
}

func VeryVerbose(msg string) {
	if verbosityLevel < 2 {
		return
	}
	printLine(os.Stdout, "TRACE", colorBlue, msg)
}

func VeryVerboseList(title string, items []string, max int) {
	if verbosityLevel < 2 || len(items) == 0 {
		return
	}
	if max <= 0 {
		max = 10
	}

	VeryVerbose(fmt.Sprintf("%s (%d)", title, len(items)))
	limit := len(items)
	if limit > max {
		limit = max
	}
	for i := 0; i < limit; i++ {
		VeryVerbose("  - " + items[i])
	}
	if len(items) > max {
		VeryVerbose(fmt.Sprintf("  ... +%d more", len(items)-max))
	}
}

func Warn(msg string) {
	printLine(os.Stderr, "WARN", colorYellow, msg)
}

func Error(msg string) {
	printLine(os.Stderr, "ERROR", colorRed, msg)
}

func Success(msg string) {
	printLine(os.Stdout, "DONE", colorGreen, msg)
}

func ReplaceLastLine(msg string) {
	if enableColor {
		fmt.Print("\033[1A\033[2K")
	}
	Info(msg)
}

func Exit(code int) {
	os.Exit(code)
}

func SetVerbosity(level int) {
	switch {
	case level < 0:
		verbosityLevel = 0
	case level > 2:
		verbosityLevel = 2
	default:
		verbosityLevel = level
	}
}

func VerbosityLevel() int {
	return verbosityLevel
}

func SemverLabel(kind string) string {
	upper := strings.ToUpper(strings.TrimSpace(kind))
	lower := strings.ToLower(upper)
	if upper == "" || !enableColor {
		return lower
	}

	var color string
	switch upper {
	case "PATCH":
		color = colorGreen
	case "MINOR":
		color = colorOrange
	case "MAJOR":
		color = colorRed
	default:
		return lower
	}

	return color + lower + colorReset
}

func PrimaryText(text string) string {
	if !enableColor {
		return text
	}
	return colorWhite + text + colorReset
}

func SecondaryText(text string) string {
	if !enableColor {
		return text
	}
	return colorGray + text + colorReset
}

func AccentText(text string) string {
	if !enableColor {
		return text
	}
	return colorSteel + text + colorReset
}

func printLine(stream *os.File, level, levelColor, msg string) {
	if !enableColor {
		fmt.Fprintf(stream, "[%s] %s\n", level, msg)
		return
	}
	fmt.Fprintf(stream, "%s%s[%s]%s %s%s%s\n", colorBold, levelColor, level, colorReset, colorDim, msg, colorReset)
}

func supportsColor() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if term := os.Getenv("TERM"); term == "" || term == "dumb" {
		return false
	}
	stdoutInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (stdoutInfo.Mode() & os.ModeCharDevice) != 0
}
