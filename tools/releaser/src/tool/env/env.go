package env

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"releaser/tool/output"
)

func Load() error {
	if _, err := os.Stat(".env"); err == nil {
		file, err := os.Open(".env")
		if err != nil {
			return fmt.Errorf("Failed to read .env: %w", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if strings.HasPrefix(line, "export ") {
				line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
			}
			key, val, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			key = strings.TrimSpace(key)
			val = strings.TrimSpace(val)
			val = strings.Trim(val, "\"'")
			if key == "" {
				continue
			}
			_ = os.Setenv(key, val)
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("Failed to read .env: %w", err)
		}
		return nil
	}

	output.Warn("No .env file is present, using 1PasswordCli")
	return nil
}

func DetectBaseDir() string {
	if stat, err := os.Stat("src"); err == nil && stat.IsDir() {
		return "src"
	}
	if stat, err := os.Stat("laravel"); err == nil && stat.IsDir() {
		return "laravel"
	}
	return "."
}

func ReadTokenFrom1Password() (string, error) {
	if _, err := exec.LookPath("op"); err != nil {
		return "", errors.New("1Password CLI (op) not found in PATH")
	}

	cmd := exec.Command("op", "read", "op://Private/GitHub Personal Access Token Studio/token")
	out, err := cmd.CombinedOutput()
	if err != nil {
		outputText := strings.TrimSpace(string(out))
		if outputText == "" {
			return "", fmt.Errorf("Failed to read GITHUB_TOKEN from 1Password: %w", err)
		}
		return "", fmt.Errorf("Failed to read GITHUB_TOKEN from 1Password: %w: %s", err, outputText)
	}
	return strings.TrimSpace(string(out)), nil
}
