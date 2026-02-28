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
	cfg := &shared.Config{}

	if err := cli.ParseArgs(cfg, os.Args[1:], os.Args[0]); err != nil {
		output.Warn(err.Error())
		cli.Usage(os.Args[0])
		output.Exit(1)
	}

	if cfg.TypeSet && !cli.IsValidType(cfg.Type) {
		output.Warn("Invalid release type: " + cfg.Type)
		output.Exit(1)
	}

	if err := env.Load(); err != nil {
		output.Warn(err.Error())
		output.Exit(1)
	}

	if cfg.BaseDir == "" {
		cfg.BaseDir = env.DetectBaseDir()
	}

	cfg.Token = os.Getenv("GITHUB_TOKEN")
	if cfg.Token == "" {
		if token, err := env.ReadTokenFrom1Password(); err == nil {
			cfg.Token = token
		} else {
			output.Warn(err.Error())
		}
	}

	if cfg.Token == "" {
		output.Warn("GITHUB_TOKEN is required and could not be loaded from .env or 1Password")
		output.Exit(1)
	}

	if !gitops.IsGitRepo(cfg.BaseDir) {
		output.Warn(fmt.Sprintf("'%s' is not a git working tree, yon can set RELEASER_BASE_DIR your .env file", cfg.BaseDir))
		output.Exit(1)
	}

	if _, err := exec.LookPath("git"); err != nil {
		output.Warn("Required command 'git' not found in PATH")
		output.Exit(1)
	}

	if err := gitops.CheckUncommittedChanges(cfg); err != nil {
		output.Exit(1)
	}
	if err := gitops.GetRepository(cfg); err != nil {
		output.Exit(1)
	}
	if err := githubapi.GetCurrentVersion(cfg); err != nil {
		output.Exit(1)
	}
	if !cfg.TypeSet {
		if err := releasetype.Detect(cfg); err != nil {
			output.Exit(1)
		}
	} else {
		output.Info("Skipping auto-detect; using provided release type: " + cfg.Type)
	}
	if err := version.Bump(cfg); err != nil {
		output.Exit(1)
	}
	if err := release.CreateTag(cfg); err != nil {
		output.Exit(1)
	}
	if err := release.BuildChanges(cfg); err != nil {
		output.Exit(1)
	}
	if err := githubapi.CreateRelease(cfg); err != nil {
		output.Exit(1)
	}

	output.Success("Created GitHub release: " + cfg.Release)
}
