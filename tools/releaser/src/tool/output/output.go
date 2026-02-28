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
	colorRed    = "\033[31m"
)

var enableColor = supportsColor()

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
