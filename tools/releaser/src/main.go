package main

import (
	"fmt"
	"os"
	"os/exec"

	"releaser/tool/cli"
	"releaser/tool/env"
	"releaser/tool/githubapi"
	"releaser/tool/gitops"
	"releaser/tool/output"
	"releaser/tool/release"
	"releaser/tool/releasetype"
	"releaser/tool/shared"
	"releaser/tool/version"
)

func main() {
	if err := run(os.Args); err != nil {
		output.Error(err.Error())
		output.Exit(1)
	}
}

func run(args []string) error {
	cfg := &shared.Config{
		Follow: true,
	}

	if err := parseAndValidateArgs(cfg, args); err != nil {
		return err
	}
	output.SetVerbosity(cfg.Verbosity)
	if cfg.Verbosity == 2 {
		output.Verbose("Very verbose logging enabled")
	} else if cfg.Verbosity == 1 {
		output.Verbose("Verbose logging enabled")
	}
	output.Verbose("CLI args parsed successfully")
	if err := prepareEnvironment(cfg); err != nil {
		return err
	}
	if err := runReleaseFlow(cfg); err != nil {
		return err
	}

	output.Success("Created GitHub release: " + cfg.Release)

	if cfg.Follow {
		if err := githubapi.FollowReleaseWorkflow(cfg); err != nil {
			output.Warn("Follow mode failed: " + err.Error())
		}
	}

	return nil
}

func parseAndValidateArgs(cfg *shared.Config, args []string) error {
	if err := cli.ParseArgs(cfg, args[1:], args[0]); err != nil {
		output.Warn(err.Error())
		cli.Usage(args[0])
		return err
	}
	if cfg.TypeSet && !cli.IsValidType(cfg.Type) {
		output.Warn("Invalid release type: " + cfg.Type)
		return fmt.Errorf("invalid release type: %s", cfg.Type)
	}
	return nil
}

func prepareEnvironment(cfg *shared.Config) error {
	output.Verbose("Loading environment variables")
	if err := env.Load(); err != nil {
		output.Warn(err.Error())
		return err
	}
	if cfg.BaseDir == "" {
		cfg.BaseDir = env.DetectBaseDir()
	}
	output.Verbose("Base directory resolved to: " + cfg.BaseDir)

	cfg.Token = os.Getenv("GITHUB_TOKEN")
	if cfg.Token == "" {
		output.Verbose("GITHUB_TOKEN not found in environment; trying 1Password")
		if token, err := env.ReadTokenFrom1Password(); err == nil {
			cfg.Token = token
			output.Verbose("GITHUB_TOKEN loaded from 1Password")
		} else {
			output.Warn(err.Error())
		}
	} else {
		output.Verbose("GITHUB_TOKEN loaded from environment")
	}
	if cfg.Token == "" {
		output.Warn("GITHUB_TOKEN is required and could not be loaded from .env or 1Password")
		return fmt.Errorf("missing GITHUB_TOKEN")
	}
	return nil
}

func runReleaseFlow(cfg *shared.Config) error {
	output.Verbose("Starting release flow")
	if !gitops.IsGitRepo(cfg.BaseDir) {
		output.Warn(fmt.Sprintf("'%s' is not a git working tree, yon can set RELEASER_BASE_DIR your .env file", cfg.BaseDir))
		return fmt.Errorf("not a git repository: %s", cfg.BaseDir)
	}
	if _, err := exec.LookPath("git"); err != nil {
		output.Warn("Required command 'git' not found in PATH")
		return err
	}

	for _, step := range []struct {
		name string
		fn   func(*shared.Config) error
	}{
		{name: "CheckUncommittedChanges", fn: gitops.CheckUncommittedChanges},
		{name: "GetRepository", fn: gitops.GetRepository},
		{name: "GetCurrentVersion", fn: githubapi.GetCurrentVersion},
	} {
		output.Verbose("Running preflight step: " + step.name)
		if err := step.fn(cfg); err != nil {
			return err
		}
	}

	if cfg.TypeSet {
		output.Info("Skipping auto-detect; using provided release type: " + cfg.Type)
	} else {
		if err := releasetype.Detect(cfg); err != nil {
			return err
		}
	}

	for _, step := range []struct {
		name string
		fn   func(*shared.Config) error
	}{
		{name: "VersionBump", fn: version.Bump},
		{name: "CreateTag", fn: release.CreateTag},
		{name: "BuildChanges", fn: release.BuildChanges},
		{name: "CreateRelease", fn: githubapi.CreateRelease},
	} {
		output.Verbose("Running release step: " + step.name)
		if err := step.fn(cfg); err != nil {
			return err
		}
	}

	return nil
}
