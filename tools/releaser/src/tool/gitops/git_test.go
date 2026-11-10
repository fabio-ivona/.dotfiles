package gitops

import (
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"releaser/tool/shared"
)

func fakeGit(t *testing.T, script string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "git"), []byte("#!/bin/sh\n"+script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return dir
}

func TestRemoteAuthenticationFallback(t *testing.T) {
	for _, args := range [][]string{
		{"ls-remote", "--tags", "origin", "refs/tags/1.0.19"},
		{"push"},
		{"push", "origin", "1.0.19"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			t.Setenv("GIT_CONFIG_COUNT", "1")
			t.Setenv("GIT_CONFIG_KEY_0", "test.inherited")
			t.Setenv("GIT_CONFIG_VALUE_0", "preserved")
			dir := fakeGit(t, `
if [ "$GIT_CONFIG_COUNT" = 1 ]; then
  echo 'git@github.com: Permission denied (publickey).' >&2
  exit 128
fi
[ "$GIT_CONFIG_COUNT" = 6 ] || exit 10
[ "$GIT_CONFIG_KEY_0" = test.inherited ] || exit 11
[ "$GIT_CONFIG_VALUE_0" = preserved ] || exit 12
[ "$GIT_CONFIG_KEY_1" = url.https://github.com/.insteadOf ] || exit 13
[ "$GIT_CONFIG_VALUE_1" = git@github.com: ] || exit 14
[ "$GIT_CONFIG_VALUE_2" = ssh://git@github.com/ ] || exit 15
[ "$GIT_CONFIG_KEY_4" = http.https://github.com/.extraheader ] || exit 16
[ "$GIT_CONFIG_VALUE_4" = "$EXPECTED_AUTH" ] || exit 17
[ "$GIT_CONFIG_VALUE_5" = '' ] || exit 18
[ "$GIT_TERMINAL_PROMPT" = 0 ] || exit 19
printf '%s\n' "$@"
`)
			t.Setenv("EXPECTED_AUTH", "AUTHORIZATION: basic "+base64.StdEncoding.EncodeToString([]byte("x-access-token:test-token")))
			out, err := RunRemote(&shared.Config{BaseDir: dir, Token: "test-token"}, args...)
			if err != nil {
				t.Fatal(err)
			}
			want := strings.Join(append([]string{"-C", dir}, args...), "\n") + "\n"
			if out != want {
				t.Fatalf("unexpected arguments: %q", out)
			}
			if os.Getenv("GIT_CONFIG_COUNT") != "1" {
				t.Fatal("fallback changed parent environment")
			}
		})
	}
}

func TestRemoteDoesNotRetryOtherResults(t *testing.T) {
	for _, tc := range []struct {
		name, script, token string
		wantErr             bool
	}{
		{"success", "echo ok", "test-token", false},
		{"network failure", "echo 'Could not resolve host' >&2; exit 128", "test-token", true},
		{"missing token", "echo 'Permission denied (publickey)' >&2; exit 128", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := fakeGit(t, "echo attempt >> \"$ATTEMPTS\"\n"+tc.script)
			path := filepath.Join(dir, "attempts")
			t.Setenv("ATTEMPTS", path)
			_, err := RunRemote(&shared.Config{BaseDir: dir, Token: tc.token}, "push")
			if (err != nil) != tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			attempts, err := os.ReadFile(path)
			if err != nil || string(attempts) != "attempt\n" {
				t.Fatal("command should run exactly once")
			}
		})
	}
}

func TestFallbackRewritesGitHubURLWithoutChangingOrigin(t *testing.T) {
	git, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	for _, origin := range []string{"git@github.com:owner/repo.git", "ssh://git@github.com/owner/repo.git"} {
		t.Run(origin, func(t *testing.T) {
			dir := t.TempDir()
			for _, args := range [][]string{{"init", dir}, {"-C", dir, "remote", "add", "origin", origin}} {
				if out, err := exec.Command(git, args...).CombinedOutput(); err != nil {
					t.Fatalf("git setup: %v: %s", err, out)
				}
			}
			t.Setenv("REAL_GIT", git)
			t.Setenv("GIT_CONFIG_COUNT", "0")
			fakeGit(t, `
if [ "$GIT_CONFIG_COUNT" = 0 ]; then
  echo 'Permission denied (publickey)' >&2
  exit 128
fi
exec "$REAL_GIT" "$@"
`)
			out, err := RunRemote(&shared.Config{BaseDir: dir, Token: "test-token"}, "remote", "get-url", "origin")
			if err != nil || strings.TrimSpace(out) != "https://github.com/owner/repo.git" {
				t.Fatalf("URL rewrite failed: %q, %v", out, err)
			}
			outBytes, err := exec.Command(git, "-C", dir, "config", "--get", "remote.origin.url").CombinedOutput()
			if err != nil || strings.TrimSpace(string(outBytes)) != origin {
				t.Fatal("origin configuration was changed")
			}
		})
	}
}

func TestFallbackRedactsCredentialsOnError(t *testing.T) {
	t.Setenv("GIT_CONFIG_COUNT", "0")
	dir := fakeGit(t, `
if [ "$GIT_CONFIG_COUNT" = 0 ]; then
  echo 'Permission denied (publickey)' >&2
else
  echo "$GIT_CONFIG_VALUE_3 test-token" >&2
fi
exit 128
`)
	out, err := RunRemote(&shared.Config{BaseDir: dir, Token: "test-token"}, "push")
	encoded := base64.StdEncoding.EncodeToString([]byte("x-access-token:test-token"))
	if err == nil || strings.Contains(out+err.Error(), "test-token") || strings.Contains(out+err.Error(), encoded) {
		t.Fatal("fallback must return a redacted error")
	}
}
