package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
)

func TestParseIssue(t *testing.T) {
	number, repo, err := parseIssue("https://github.com/dapi/start-issue/issues/34", "")
	if err != nil || number != "34" || repo != "dapi/start-issue" {
		t.Fatalf("got %q %q %v", number, repo, err)
	}
}

func TestParseTracksWorktreeDirectorySource(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("START_ISSUE_WORKTREE_DIR", "")

	o, err := parse(nil)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := o.worktreeDirSource, "built-in default"; got != want {
		t.Fatalf("default worktree directory source = %q, want %q", got, want)
	}

	t.Setenv("START_ISSUE_WORKTREE_DIR", "/env-worktrees")
	o, err = parse(nil)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := o.worktreeDirSource, "START_ISSUE_WORKTREE_DIR"; got != want {
		t.Fatalf("environment worktree directory source = %q, want %q", got, want)
	}

	o, err = parse([]string{"--worktree-dir", "/cli-worktrees"})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := o.worktreeDirSource, "CLI"; got != want {
		t.Fatalf("CLI worktree directory source = %q, want %q", got, want)
	}
}

func TestUserHomeDirRejectsUnavailableOrRelativeHome(t *testing.T) {
	t.Setenv("HOME", "")
	if runtime.GOOS != "windows" {
		if _, err := userHomeDir(); err == nil {
			t.Fatal("empty HOME returned a usable home directory")
		}
	}

	t.Setenv("HOME", "relative-home")
	if runtime.GOOS != "windows" {
		if _, err := userHomeDir(); err == nil {
			t.Fatal("relative HOME returned a usable home directory")
		}
	}
}

func TestParseRequiresHomeOnlyForDefaultWorktreeDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows derives its home directory from USERPROFILE")
	}
	t.Setenv("HOME", "")
	t.Setenv("START_ISSUE_WORKTREE_DIR", "")

	if _, err := parse([]string{"1"}); err == nil || !strings.Contains(err.Error(), "set --worktree-dir or START_ISSUE_WORKTREE_DIR") {
		t.Fatalf("default worktree directory error = %v", err)
	}
	if o, err := parse([]string{"1", "--worktree-dir", t.TempDir()}); err != nil || o.worktreeDirSource != "CLI" {
		t.Fatalf("explicit worktree directory = %#v, %v", o, err)
	}
}

func TestResolversDoNotReadRelativeUserConfigWhenHomeIsUnavailable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows derives its home directory from USERPROFILE")
	}
	wd := t.TempDir()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(previous) }()
	if err := os.Chdir(wd); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(".config", "start-issue"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(".config", "start-issue", "agent"), []byte("codex\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", "")

	agent, source, err := resolveAgent("", "")
	if err != nil || agent != "claude" || source != "built-in default" {
		t.Fatalf("resolveAgent = %q, %q, %v", agent, source, err)
	}
}

func TestResolvePromptReportsMissingPromptFile(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing-prompt.md")
	for _, test := range []struct {
		name string
		o    options
		env  bool
	}{
		{name: "CLI", o: options{promptFile: missing}},
		{name: "environment", env: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("START_ISSUE_PROMPT", "")
			if test.env {
				t.Setenv("START_ISSUE_PROMPT_FILE", missing)
			} else {
				t.Setenv("START_ISSUE_PROMPT_FILE", "")
			}

			_, _, _, _, err := resolvePrompt(t.TempDir(), "none", test.o)
			if err == nil || err.Error() != "Prompt file not found: "+missing {
				t.Fatalf("resolvePrompt error = %v, want public missing-file error", err)
			}
			if !errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("resolvePrompt error = %v, want errors.Is(err, fs.ErrNotExist)", err)
			}
		})
	}
}

func TestResolvePromptStripsTrailingNewlinesFromPromptFiles(t *testing.T) {
	root, home, promptFile := t.TempDir(), t.TempDir(), filepath.Join(t.TempDir(), "prompt.md")
	const content = "Prompt {ISSUE_URL}\n\n"
	if err := os.WriteFile(promptFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".start-issue"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".start-issue", "prompt.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, ".config", "start-issue"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".config", "start-issue", "prompt.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)

	for _, test := range []struct {
		name string
		root string
		o    options
		env  bool
	}{
		{name: "CLI", root: t.TempDir(), o: options{promptFile: promptFile}},
		{name: "environment", root: t.TempDir(), env: true},
		{name: "project config", root: root},
		{name: "user config", root: t.TempDir()},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("START_ISSUE_PROMPT", "")
			if test.env {
				t.Setenv("START_ISSUE_PROMPT_FILE", promptFile)
			} else {
				t.Setenv("START_ISSUE_PROMPT_FILE", "")
			}

			got, _, _, _, err := resolvePrompt(test.root, "codex", test.o)
			if err != nil || got != "Prompt {ISSUE_URL}" {
				t.Fatalf("resolvePrompt() = %q, %v; want prompt without trailing newlines", got, err)
			}
		})
	}
}

func TestBranchName(t *testing.T) {
	if got := branchName("34", "Fix broken output!", "bug"); got != "fix/issue-34-fix-broken-output" {
		t.Fatalf("got %q", got)
	}
}

func TestBranchNameMatchesFastShellRules(t *testing.T) {
	tests := []struct {
		title, labels, want string
	}{
		{"[brief] Исправить ЦАП", "urgent, bug", "hotfix/issue-34-ispravit-tsap"},
		{"Documentation tidy", "documentation", "docs/issue-34-documentation-tidy"},
		{"Fix broken output", "Bug", "feature/issue-34-fix-broken-output"},
		{"", "", "feature/issue-34-work"},
	}
	for _, test := range tests {
		if got := branchName("34", test.title, test.labels); got != test.want {
			t.Errorf("branchName(%q, %q) = %q, want %q", test.title, test.labels, got, test.want)
		}
	}
}

func TestSlugifyTruncatesBeforeTrimmingSeparators(t *testing.T) {
	title := "😀" + strings.Repeat("a", 39) + " trailing"
	if got, want := slugify(title), strings.Repeat("a", 39); got != want {
		t.Fatalf("slugify(%q) = %q, want %q", title, got, want)
	}
}

func TestRender(t *testing.T) {
	if got := render("{REPO} #{ISSUE_NUMBER}", map[string]string{"REPO": "dapi/start-issue", "ISSUE_NUMBER": "34"}); got != "dapi/start-issue #34" {
		t.Fatalf("got %q", got)
	}
}

func TestRenderPreservesOrderedExpansionOfIssueData(t *testing.T) {
	in := issue{Title: "Keep {REPO}", Body: "and {ISSUE_NUMBER}"}
	got := renderIssuePrompt("{ISSUE_URL}; {ISSUE_NUMBER}; {ISSUE_TITLE}; {ISSUE_BODY}; {ISSUE_LABELS}; {REPO}; {BRANCH_NAME}; {WORKTREE_PATH}; {BASE_BRANCH}", "https://github.com/dapi/start-issue/issues/34", "34", in, []string{"bug"}, "dapi/start-issue", "feature/issue-34", "/tmp/reused", "main")
	want := "https://github.com/dapi/start-issue/issues/34; 34; Keep dapi/start-issue; and {ISSUE_NUMBER}; bug; dapi/start-issue; feature/issue-34; /tmp/reused; main"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestIssueUnmarshalNormalizesNullBody(t *testing.T) {
	var in issue
	if err := json.Unmarshal([]byte(`{"title":"No description","body":null,"labels":[{"name":"bug"}]}`), &in); err != nil {
		t.Fatal(err)
	}
	if in.Body != "" {
		t.Fatalf("Body = %q, want empty string", in.Body)
	}
	if got := renderIssuePrompt("{ISSUE_BODY}", "", "", in, nil, "", "", "", ""); got != "" {
		t.Fatalf("rendered null body = %q, want empty string", got)
	}
}

func TestDetectRepoTrimsNewlineBeforeGitSuffix(t *testing.T) {
	bin := t.TempDir()
	writeExecutable(t, filepath.Join(bin, "git"), "#!/bin/sh\nprintf '%s\\n' 'https://github.com/dapi/start-issue.git'\n")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	got, err := detectRepo()
	if err != nil || got != "dapi/start-issue" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestDryRunLaunchDoesNotExecuteAgent(t *testing.T) {
	bin := t.TempDir()
	marker := filepath.Join(t.TempDir(), "codex-ran")
	writeExecutable(t, filepath.Join(bin, "codex"), "#!/bin/sh\ntouch '"+marker+"'\n")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	if err := launchSelected(options{dryRun: true}, "codex", "", t.TempDir(), "prompt"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("dry-run executed codex: %v", err)
	}
}

func TestDryRunDeleteAndRecreatePreservesExistingBranchAndWorktree(t *testing.T) {
	home, bin, root, worktrees := t.TempDir(), t.TempDir(), t.TempDir(), t.TempDir()
	worktree := filepath.Join(worktrees, "feature", "issue-1-add-login-button")
	worktreeMarker := filepath.Join(worktree, "uncommitted-work")
	branchMarker := filepath.Join(t.TempDir(), "branch-exists")
	gitLog := filepath.Join(t.TempDir(), "git-mutations.log")
	if err := os.MkdirAll(filepath.Join(home, ".config", "start-issue"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(worktree, 0755); err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{worktreeMarker, branchMarker} {
		if err := os.WriteFile(marker, []byte("must remain"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	writeExecutable(t, filepath.Join(bin, "git"), fmt.Sprintf(`#!/bin/sh
case "$1 $2 $3" in
  "rev-parse --show-toplevel ") printf '%%s\n' %q ;;
  "remote get-url origin") printf '%%s\n' 'git@github.com:owner/repo.git' ;;
  "show-ref --verify --quiet") [ "$4" = "refs/heads/feature/issue-1-add-login-button" ] && exit 0; exit 1 ;;
  "worktree list --porcelain") printf 'worktree %%s\nbranch refs/heads/master\n\nworktree %%s\nbranch refs/heads/feature/issue-1-add-login-button\n' %q %q ;;
  "worktree remove --force") printf '%%s\n' "$*" >> "$START_ISSUE_GIT_LOG"; rm -rf "$4" ;;
esac
if [ "$1 $2" = "branch -D" ]; then
  printf '%%s\n' "$*" >> "$START_ISSUE_GIT_LOG"
  rm -f "$START_ISSUE_BRANCH_MARKER"
fi
`, root, root, worktree))
	writeExecutable(t, filepath.Join(bin, "gh"), `#!/bin/sh
if [ "$1 $2" = "auth status" ]; then exit 0; fi
if [ "$1" = api ]; then
  printf '%s\n' '{"title":"Add login button","body":"","labels":[]}'
fi
`)
	t.Setenv("HOME", home)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("START_ISSUE_GIT_LOG", gitLog)
	t.Setenv("START_ISSUE_BRANCH_MARKER", branchMarker)

	output := captureStdout(t, func() {
		err := runWithReader(options{
			issue:       "1",
			repo:        "owner/repo",
			base:        "master",
			worktreeDir: worktrees,
			agent:       "none",
			dryRun:      true,
			noInit:      true,
		}, bufio.NewReader(strings.NewReader("3\n")))
		if err != nil {
			t.Fatal(err)
		}
	})
	for _, marker := range []string{worktreeMarker, branchMarker} {
		if _, err := os.Stat(marker); err != nil {
			t.Fatalf("dry-run removed %s: %v", marker, err)
		}
	}
	if mutations, err := os.ReadFile(gitLog); err == nil && len(mutations) > 0 {
		t.Fatalf("dry-run ran destructive git commands:\n%s", mutations)
	}
	for _, want := range []string{
		"[DRY-RUN] Would remove worktree: " + worktree,
		"[DRY-RUN] Would delete branch: feature/issue-1-add-login-button",
		"[DRY-RUN] Would run: git worktree add -b feature/issue-1-add-login-button",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("dry-run output missing %q:\n%s", want, output)
		}
	}
}

func TestSuffixWorktreeUsesSuffixedBranch(t *testing.T) {
	bin := t.TempDir()
	writeExecutable(t, filepath.Join(bin, "git"), `#!/bin/sh
case "$4" in
  refs/heads/feature/issue-34-original|refs/heads/feature/issue-34-original-v2) exit 0 ;;
  *) exit 1 ;;
esac
`)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	branch := nextSuffixedBranch("feature/issue-34-original")
	if branch != "feature/issue-34-original-v3" {
		t.Fatalf("got branch %q", branch)
	}
	if got, want := worktreePath("/worktrees", branch, false), "/worktrees/feature/issue-34-original-v3"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRemoveWorktreeAndBranchRemovesStaleRegistration(t *testing.T) {
	bin, log := t.TempDir(), filepath.Join(t.TempDir(), "git.log")
	writeExecutable(t, filepath.Join(bin, "git"), `#!/bin/sh
printf '%s\n' "$*" >> "$START_ISSUE_GIT_LOG"
if [ "$1 $2" = "worktree remove" ] && [ "$START_ISSUE_REMOVE_WORKTREE_FAIL" = 1 ]; then
  exit 1
fi
`)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("START_ISSUE_GIT_LOG", log)

	stale := filepath.Join(t.TempDir(), "externally-deleted-worktree")
	if err := removeWorktreeAndBranch(stale, "feature/issue-34-stale"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"worktree remove --force " + stale,
		"branch -D feature/issue-34-stale",
	} {
		if !strings.Contains(string(got), want) {
			t.Fatalf("git calls missing %q:\n%s", want, got)
		}
	}
}

func TestRemoveWorktreeAndBranchPrunesAfterFailedStaleRemoval(t *testing.T) {
	bin, log := t.TempDir(), filepath.Join(t.TempDir(), "git.log")
	writeExecutable(t, filepath.Join(bin, "git"), `#!/bin/sh
printf '%s\n' "$*" >> "$START_ISSUE_GIT_LOG"
if [ "$1 $2" = "worktree remove" ]; then
  exit 1
fi
`)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("START_ISSUE_GIT_LOG", log)

	stale := filepath.Join(t.TempDir(), "externally-deleted-worktree")
	if err := removeWorktreeAndBranch(stale, "feature/issue-34-stale"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "worktree prune") {
		t.Fatalf("failed stale removal did not prune registration:\n%s", got)
	}
}

func TestRemoveWorktreeAndBranchRefusesPrimaryWorktree(t *testing.T) {
	bin, log, primary := t.TempDir(), filepath.Join(t.TempDir(), "git.log"), t.TempDir()
	marker := filepath.Join(primary, "keep")
	if err := os.WriteFile(marker, []byte("must not be deleted"), 0644); err != nil {
		t.Fatal(err)
	}
	writeExecutable(t, filepath.Join(bin, "git"), fmt.Sprintf(`#!/bin/sh
printf '%%s\n' "$*" >> "$START_ISSUE_GIT_LOG"
case "$1 $2 $3" in
  "worktree list --porcelain") printf 'worktree %s\nbranch refs/heads/main\n' %q ;;
  "rev-parse --show-toplevel") printf '%%s\n' %q ;;
esac
	`, primary, primary, primary))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("START_ISSUE_GIT_LOG", log)

	err := removeWorktreeAndBranch(primary, "main")
	if err == nil || !strings.Contains(err.Error(), "not a removable linked worktree") {
		t.Fatalf("remove primary worktree error = %v", err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("primary worktree content was deleted: %v", err)
	}
	got, _ := os.ReadFile(log)
	if strings.Contains(string(got), "worktree remove") || strings.Contains(string(got), "branch -D") {
		t.Fatalf("primary worktree removal invoked git mutation:\n%s", got)
	}
}

func TestRemoveWorktreeAndBranchNeverFallsBackToRawDeletion(t *testing.T) {
	bin, log, primary, linked := t.TempDir(), filepath.Join(t.TempDir(), "git.log"), t.TempDir(), t.TempDir()
	marker := filepath.Join(linked, "keep")
	if err := os.WriteFile(marker, []byte("must not be deleted"), 0644); err != nil {
		t.Fatal(err)
	}
	writeExecutable(t, filepath.Join(bin, "git"), fmt.Sprintf(`#!/bin/sh
printf '%%s\n' "$*" >> "$START_ISSUE_GIT_LOG"
case "$1 $2 $3" in
  "worktree list --porcelain") printf 'worktree %s\nbranch refs/heads/main\nworktree %s\nbranch refs/heads/feature/issue-34\n' %q %q ;;
  "rev-parse --show-toplevel") printf '%%s\n' %q ;;
esac
if [ "$1 $2" = "worktree remove" ]; then exit 1; fi
`, primary, linked, primary, linked, primary))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("START_ISSUE_GIT_LOG", log)

	err := removeWorktreeAndBranch(linked, "feature/issue-34")
	if err == nil || !strings.Contains(err.Error(), "refusing to delete worktree path") {
		t.Fatalf("failed linked worktree removal error = %v", err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("linked worktree content was deleted: %v", err)
	}
	got, _ := os.ReadFile(log)
	if strings.Contains(string(got), "worktree prune") || strings.Contains(string(got), "branch -D") {
		t.Fatalf("failed removal mutated stale registration or branch:\n%s", got)
	}
}

func TestReleaseAssetName(t *testing.T) {
	for _, test := range []struct {
		goos, goarch, want string
	}{
		{"linux", "amd64", "start-issue-linux-amd64"},
		{"linux", "arm64", "start-issue-linux-arm64"},
		{"darwin", "amd64", "start-issue-darwin-amd64"},
		{"darwin", "arm64", "start-issue-darwin-arm64"},
		{"windows", "amd64", "start-issue-windows-amd64.exe"},
	} {
		got, err := releaseAssetName(test.goos, test.goarch)
		if err != nil || got != test.want {
			t.Fatalf("releaseAssetName(%s, %s) = %q, %v; want %q", test.goos, test.goarch, got, err, test.want)
		}
	}
	for _, test := range []struct{ goos, goarch string }{{"linux", "386"}, {"freebsd", "amd64"}, {"windows", "arm64"}} {
		if got, err := releaseAssetName(test.goos, test.goarch); err == nil || got != "" || !strings.Contains(err.Error(), "unsupported release platform") {
			t.Fatalf("releaseAssetName(%s, %s) = %q, %v; want unsupported platform error", test.goos, test.goarch, got, err)
		}
	}
}

func TestPathConflictChoiceRejectsUnregisteredDirectoryWithoutDeleteOption(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ordinary-directory")
	output := captureStdout(t, func() {
		_, _ = pathConflictChoice(path, "", false, bufio.NewReader(strings.NewReader("2\n")))
	})
	_, err := pathConflictChoice(path, "", false, bufio.NewReader(strings.NewReader("2\n")))
	if err == nil || !strings.Contains(err.Error(), "Move or remove the directory manually") {
		t.Fatalf("unregistered path conflict error = %v", err)
	}
	if strings.Contains(output, "Delete and recreate") || strings.Contains(output, "Choice:") {
		t.Fatalf("unregistered path conflict offered deletion:\n%s", output)
	}
}

func TestConflictChoicesReportWaitingForInput(t *testing.T) {
	branchOutput := captureStdout(t, func() {
		_ = branchConflictChoice("feature/issue-1-demo", "", bufio.NewReader(strings.NewReader("0\n")))
	})
	if !strings.Contains(branchOutput, "Waiting for input: branch already exists") {
		t.Fatalf("branch conflict did not report waiting for input:\n%s", branchOutput)
	}

	pathOutput := captureStdout(t, func() {
		_, _ = pathConflictChoice(t.TempDir(), "refs/heads/feature/issue-1-demo", true, bufio.NewReader(strings.NewReader("0\n")))
	})
	if !strings.Contains(pathOutput, "Waiting for input: worktree path already exists") {
		t.Fatalf("path conflict did not report waiting for input:\n%s", pathOutput)
	}
}

func TestDetachedWorktreeIsRegisteredAndRemovable(t *testing.T) {
	bin, log, primary, detached := t.TempDir(), filepath.Join(t.TempDir(), "git.log"), t.TempDir(), t.TempDir()
	writeExecutable(t, filepath.Join(bin, "git"), fmt.Sprintf(`#!/bin/sh
printf '%%s\n' "$*" >> "$START_ISSUE_GIT_LOG"
case "$1 $2 $3" in
  "worktree list --porcelain") printf 'worktree %%s\nbranch refs/heads/main\n\nworktree %%s\nHEAD deadbeef\ndetached\n' %q %q ;;
  "rev-parse --show-toplevel") printf '%%s\n' %q ;;
esac
`, primary, detached, primary))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("START_ISSUE_GIT_LOG", log)

	branch, registered := worktreeRegistration(detached)
	if !registered || branch != "" {
		t.Fatalf("worktreeRegistration(detached) = %q, %t; want empty branch and registered", branch, registered)
	}
	output := captureStdout(t, func() {
		choice, err := pathConflictChoice(detached, branch, registered, bufio.NewReader(strings.NewReader("2\n")))
		if err != nil || choice != "2" {
			t.Fatalf("detached path conflict choice = %q, %v", choice, err)
		}
	})
	if !strings.Contains(output, "Registered branch: detached HEAD") || !strings.Contains(output, "Delete and recreate") {
		t.Fatalf("detached worktree did not offer safe delete/recreate:\n%s", output)
	}
	if err := removeWorktreeAndBranch(detached, branch); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(log)
	if err != nil || !strings.Contains(string(got), "worktree remove --force "+detached) {
		t.Fatalf("detached worktree was not removed through git: %q, %v", got, err)
	}
	if strings.Contains(string(got), "branch -D") {
		t.Fatalf("detached worktree removal deleted a branch:\n%s", got)
	}
}

func TestValidChecksum(t *testing.T) {
	binary := []byte("start-issue")
	checksum := sha256.Sum256(binary)
	manifest := fmt.Sprintf("%x  start-issue-darwin-arm64\n", checksum)
	if !validChecksum(binary, "start-issue-darwin-arm64", manifest) {
		t.Fatal("expected checksum to match")
	}
	if validChecksum(binary, "start-issue-linux-amd64", manifest) {
		t.Fatal("unexpected checksum match")
	}
}

func TestCompareVersions(t *testing.T) {
	for _, test := range []struct {
		left, right string
		want        int
	}{
		{left: "v1.2.0", right: "1.2.0", want: 0},
		{left: "1.2.1", right: "1.2.0", want: 1},
		{left: "1.1.9", right: "v1.2.0", want: -1},
		{left: "2.0.0-rc.1", right: "v2.0.0", want: -1},
		{left: "2.0.0-rc.2", right: "2.0.0-rc.1", want: 1},
		{left: "2.0.0-alpha", right: "2.0.0-alpha.1", want: -1},
		{left: "1.13.2-8-gabcdef", right: "v1.13.2", want: 1},
		{left: "1.13.2-8-gabcdef-dirty", right: "v1.13.2", want: 1},
		{left: "1.13.2-dirty", right: "v1.13.2", want: 1},
	} {
		if got := compareVersions(test.left, test.right); got != test.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", test.left, test.right, got, test.want)
		}
	}
}

func TestUpdateUsesConfiguredReleaseRepository(t *testing.T) {
	bin, log := t.TempDir(), filepath.Join(t.TempDir(), "gh.log")
	writeExecutable(t, filepath.Join(bin, "gh"), `#!/bin/sh
printf '%s\n' "$*" >> "$START_ISSUE_GH_LOG"
printf '%s\n' '{"tag_name":"v1.0.0"}'
`)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("START_ISSUE_GH_LOG", log)
	t.Setenv("START_ISSUE_REPOSITORY", "fork/start-issue")

	previousVersion := version
	version = "2.0.0"
	defer func() { version = previousVersion }()

	if err := updateMode(options{}); err != nil {
		t.Fatal(err)
	}
	called, err := os.ReadFile(log)
	if err != nil || strings.TrimSpace(string(called)) != "auth status\napi repos/fork/start-issue/releases/latest" {
		t.Fatalf("gh API call = %q, %v", called, err)
	}
}

func TestUpdateRequiresAuthenticatedGitHubCLI(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows update exits through its documented manual path")
	}
	bin := t.TempDir()
	writeExecutable(t, filepath.Join(bin, "gh"), "#!/bin/sh\n[ \"$1\" = auth ] && exit 1\n")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	err := updateMode(options{})
	if err == nil || err.Error() != "gh not authenticated. Run: gh auth login" {
		t.Fatalf("update authentication error = %v", err)
	}
}

func TestInstallRequiresGitHubCLI(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows installation is intentionally manual")
	}
	t.Setenv("PATH", t.TempDir())
	if err := installMode(false); err == nil || err.Error() != "gh CLI not found. Install: https://cli.github.com" {
		t.Fatalf("install GitHub CLI error = %v", err)
	}
}

func TestUpdateRequiresGitHubCLI(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows update is intentionally manual")
	}
	t.Setenv("PATH", t.TempDir())
	if err := updateMode(options{}); err == nil || err.Error() != "gh CLI not found. Install: https://cli.github.com" {
		t.Fatalf("update GitHub CLI error = %v", err)
	}
}

func TestCheckGitHubAccessReportsInstallationInstructions(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if err := checkGitHubAccess(); err == nil || err.Error() != "gh CLI not found. Install: https://cli.github.com" {
		t.Fatalf("GitHub CLI error = %v", err)
	}
}

func TestRunModeUpdateDoesNotRequireHomeDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows update exits through its documented manual path")
	}
	bin := t.TempDir()
	writeExecutable(t, filepath.Join(bin, "gh"), "#!/bin/sh\nprintf '%s\\n' '{\"tag_name\":\"v1.0.0\"}'\n")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("HOME", "")
	previousVersion := version
	version = "1.0.0"
	defer func() { version = previousVersion }()

	if err := runMode(options{mode: "update"}); err != nil {
		t.Fatalf("update unexpectedly required HOME: %v", err)
	}
}

func TestUpdateRejectsReleaseWithoutTagName(t *testing.T) {
	bin := t.TempDir()
	writeExecutable(t, filepath.Join(bin, "gh"), "#!/bin/sh\nprintf '%s\\n' '{}'\n")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	err := updateMode(options{})
	if err == nil || !strings.Contains(err.Error(), "missing tag_name") {
		t.Fatalf("update error = %v", err)
	}
}

func TestUpdateDryRunRejectsReleaseWithoutRequiredAssets(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows update exits through its documented manual path")
	}
	bin := t.TempDir()
	writeExecutable(t, filepath.Join(bin, "gh"), "#!/bin/sh\nprintf '%s\\n' '{\"tag_name\":\"v9.0.0\",\"assets\":[]}'\n")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	previousVersion := version
	version = "1.0.0"
	defer func() { version = previousVersion }()

	err := updateMode(options{dryRun: true})
	if err == nil || !strings.Contains(err.Error(), "does not contain") || !strings.Contains(err.Error(), "checksums.txt") {
		t.Fatalf("dry-run missing release assets error = %v", err)
	}
}

func TestVersionFromBuildInfoUsesGoInstallModuleVersion(t *testing.T) {
	info := &debug.BuildInfo{Main: debug.Module{Version: "v2.3.4"}}
	if got := versionFromBuildInfo(info); got != "2.3.4" {
		t.Fatalf("versionFromBuildInfo() = %q, want 2.3.4", got)
	}
	if got := versionFromBuildInfo(&debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, "9.9.9"); got != "9.9.9" {
		t.Fatalf("development version = %q, want fallback", got)
	}
	if got := versionFromBuildInfo(&debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}); got != "dev" {
		t.Fatalf("unversioned development build = %q, want dev", got)
	}
}

func TestRunningVersionUsesInjectedBuildVersion(t *testing.T) {
	previousVersion := version
	version = "v9.8.7"
	defer func() { version = previousVersion }()
	if got := runningVersion(); got != "9.8.7" {
		t.Fatalf("runningVersion() = %q, want injected version", got)
	}
}

func TestParseInitOptions(t *testing.T) {
	o, err := parse([]string{"init", "--project", "--force", "--prompt-file", "prompt.md"})
	if err != nil || o.mode != "init" || !o.project || !o.force || o.promptFile != "prompt.md" {
		t.Fatalf("unexpected options: %#v, %v", o, err)
	}
}

func TestParseRejectsEmptyOptionValues(t *testing.T) {
	for _, option := range []string{"--repo", "--base", "--worktree-dir", "--agent", "--model", "--prompt-file", "--prompt", "--command", "--prompt-output-file"} {
		t.Run(option, func(t *testing.T) {
			_, err := parse([]string{option, ""})
			if err == nil || err.Error() != option+" requires a value." {
				t.Fatalf("parse(%s, empty) error = %v", option, err)
			}
		})
	}
}

func TestInitRejectsEmptyRetainedAgentConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent")
	if err := os.WriteFile(path, []byte("# configured later\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, _, err := resolveInitAgent(path, "codex", false)
	if err == nil || !strings.Contains(err.Error(), "Agent config is empty") {
		t.Fatalf("resolveInitAgent error = %v", err)
	}
}

func TestInitRejectsEmptyRetainedModelConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "model")
	if err := os.WriteFile(path, []byte("\n# configured later\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := resolveInitModel(path, "claude-opus", false)
	if err == nil || !strings.Contains(err.Error(), "Model config is empty") {
		t.Fatalf("resolveInitModel error = %v", err)
	}
}

func TestInitRetainsExistingModelBeforeCLIModel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "model")
	if err := os.WriteFile(path, []byte("configured-model\n"), 0644); err != nil {
		t.Fatal(err)
	}
	model, err := resolveInitModel(path, "cli-model", false)
	if err != nil || model != "configured-model" {
		t.Fatalf("resolveInitModel = %q, %v", model, err)
	}
}

func TestInitValidatesCLIModel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "model")

	model, err := resolveInitModel(path, "  gpt-5.2  ", true)
	if err != nil || model != "gpt-5.2" {
		t.Fatalf("resolveInitModel trimmed model = %q, %v", model, err)
	}

	_, err = resolveInitModel(path, "   ", true)
	if err == nil || !strings.Contains(err.Error(), "Model config is empty") {
		t.Fatalf("resolveInitModel whitespace-only model error = %v", err)
	}
}

func TestResolveModelNormalizesCLIValue(t *testing.T) {
	model, source, err := resolveModel(t.TempDir(), "  gpt-5  ")
	if err != nil || model != "gpt-5" || source != "CLI" {
		t.Fatalf("resolveModel trimmed CLI model = %q, %q, %v", model, source, err)
	}

	_, source, err = resolveModel(t.TempDir(), "   ")
	if err == nil || source != "CLI" || !strings.Contains(err.Error(), "--model requires a non-empty value") {
		t.Fatalf("resolveModel whitespace-only CLI model = source %q, error %v", source, err)
	}
}

func TestResolversRejectWhitespaceOnlyEnvironmentValues(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	root := t.TempDir()

	t.Setenv("START_ISSUE_AGENT", " \t\n ")
	_, source, err := resolveAgent(root, "")
	if source != "START_ISSUE_AGENT" || err == nil || err.Error() != "Agent config is empty. Valid agents: claude, codex, kimi, pi, none." {
		t.Fatalf("resolveAgent whitespace-only environment value = source %q, error %v", source, err)
	}

	t.Setenv("START_ISSUE_AGENT", "")
	t.Setenv("START_ISSUE_MODEL", " \t\n ")
	_, source, err = resolveModel(root, "")
	if source != "START_ISSUE_MODEL" || err == nil || err.Error() != "Model config is empty. Remove the empty model config or set a value." {
		t.Fatalf("resolveModel whitespace-only environment value = source %q, error %v", source, err)
	}
}

func TestParseRejectsMultipleCommandModes(t *testing.T) {
	_, err := parse([]string{"init", "update"})
	if err == nil || !strings.Contains(err.Error(), "only one command mode") {
		t.Fatalf("got %v", err)
	}
}

func TestSelectInitDirRequiresExplicitValidScopeInRepository(t *testing.T) {
	root, home := t.TempDir(), t.TempDir()
	for _, test := range []struct {
		name, input string
		want        string
		wantErr     string
	}{
		{name: "project", input: "project\n", want: filepath.Join(root, ".start-issue")},
		{name: "user", input: "2\n", want: filepath.Join(home, ".config", "start-issue")},
		{name: "invalid", input: "maybe\n", wantErr: "Invalid init scope"},
		{name: "eof", input: "", wantErr: "No init scope selected"},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := selectInitDir(root, home, options{mode: "init"}, bufio.NewReader(strings.NewReader(test.input)))
			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("error = %v, want %q", err, test.wantErr)
				}
				return
			}
			if err != nil || got != test.want {
				t.Fatalf("got %q, %v; want %q", got, err, test.want)
			}
		})
	}
}

func TestCheckGHAuthRequiresAnAuthenticatedSession(t *testing.T) {
	bin := t.TempDir()
	writeExecutable(t, filepath.Join(bin, "gh"), "#!/bin/sh\n[ \"$1\" = auth ] && [ \"$2\" = status ] && exit 1\nexit 0\n")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	if err := checkGHAuth(); err == nil || !strings.Contains(err.Error(), "gh not authenticated") {
		t.Fatalf("got %v", err)
	}
}

func TestVerifyStagedBinaryRequiresExpectedVersion(t *testing.T) {
	staged := filepath.Join(t.TempDir(), "start-issue")
	writeExecutable(t, staged, "#!/bin/sh\nif [ \"$1\" = --version ]; then echo 'start-issue v2.1.0'; fi\n")
	if err := verifyStagedBinary(staged, "v2.1.0"); err != nil {
		t.Fatal(err)
	}
	if err := verifyStagedBinary(staged, "v2.2.0"); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("got %v", err)
	}
}

func TestInstallVerifiedUpdateKeepsCurrentExecutableWhenStagingFails(t *testing.T) {
	target := filepath.Join(t.TempDir(), "start-issue")
	writeExecutable(t, target, "#!/bin/sh\n[ \"$1\" = --version ] && echo 'start-issue v2.0.0'\n")
	wrongVersion := []byte("#!/bin/sh\n[ \"$1\" = --version ] && echo 'start-issue v2.1.1'\n")
	if err := installVerifiedUpdate(target, wrongVersion, "v2.1.0"); err == nil {
		t.Fatal("expected staged version verification error")
	}
	version, err := exec.Command(target, "--version").Output()
	if err != nil || strings.TrimSpace(string(version)) != "start-issue v2.0.0" {
		t.Fatalf("current executable changed: %q, %v", version, err)
	}
	if _, err := os.Stat(target + ".new"); !os.IsNotExist(err) {
		t.Fatalf("staged executable was not removed: %v", err)
	}
}

func TestInstallVerifiedUpdateDoesNotFollowPredictableStagingSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "start-issue")
	staging := target + ".new"
	writeExecutable(t, target, "#!/bin/sh\n[ \"$1\" = --version ] && echo 'start-issue v2.0.0'\n")
	if err := os.Symlink(target, staging); err != nil {
		t.Fatal(err)
	}
	updated := []byte("#!/bin/sh\n[ \"$1\" = --version ] && echo 'start-issue v2.1.0'\n")
	if err := installVerifiedUpdate(target, updated, "v2.1.0"); err != nil {
		t.Fatal(err)
	}
	version, err := exec.Command(target, "--version").Output()
	if err != nil || strings.TrimSpace(string(version)) != "start-issue v2.1.0" {
		t.Fatalf("updated executable = %q, %v", version, err)
	}
	info, err := os.Lstat(staging)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("predictable staging symlink was modified: %v, %v", info, err)
	}
}

func TestInstallModeKeepsTargetWhenStagedVersionDoesNotMatchRelease(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows installation is intentionally manual")
	}
	home, bin := t.TempDir(), t.TempDir()
	assetName, err := releaseAssetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Skipf("unsupported test platform: %v", err)
	}
	asset := []byte("#!/bin/sh\n[ \"$1\" = --version ] && echo 'start-issue v2.1.1'\n")
	checksum := fmt.Sprintf("%x  %s\n", sha256.Sum256(asset), assetName)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("local HTTP listener unavailable: %v", err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/asset":
			_, _ = response.Write(asset)
		case "/checksums":
			_, _ = response.Write([]byte(checksum))
		default:
			http.NotFound(response, request)
		}
	}))
	server.Listener = listener
	server.Start()
	defer server.Close()
	release := filepath.Join(t.TempDir(), "release.json")
	metadata := fmt.Sprintf(`{"tag_name":"v2.1.0","assets":[{"name":"%s","browser_download_url":"%s/asset"},{"name":"checksums.txt","browser_download_url":"%s/checksums"}]}`, assetName, server.URL, server.URL)
	if err := os.WriteFile(release, []byte(metadata), 0644); err != nil {
		t.Fatal(err)
	}
	writeExecutable(t, filepath.Join(bin, "gh"), "#!/bin/sh\ncat \"$START_ISSUE_TEST_RELEASE\"\n")
	t.Setenv("HOME", home)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("START_ISSUE_TEST_RELEASE", release)

	err = installMode(false)
	if err == nil || !strings.Contains(err.Error(), "does not match expected release") {
		t.Fatalf("install error = %v", err)
	}
	target := filepath.Join(home, ".local", "bin", "start-issue")
	if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
		t.Fatalf("invalid staged binary was installed: %v", statErr)
	}
}

func TestNormalizePromptProposalStripsOnlyOuterFences(t *testing.T) {
	got := normalizePromptProposal("```markdown\n# Prompt\n```go\nkeep\n```\n```")
	if want := "# Prompt\n```go\nkeep\n```"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got := normalizePromptProposal("# Prompt\n```\n"); got != "# Prompt\n```" {
		t.Fatalf("unpaired fence changed: %q", got)
	}
}

func TestPromptImprovementOutputPathPreservesLegacyFileNames(t *testing.T) {
	root := t.TempDir()
	for _, test := range []struct {
		name, path, want string
	}{
		{"markdown", filepath.Join(root, "prompt.md"), filepath.Join(root, "prompt.improved.md")},
		{"other extension", filepath.Join(root, "prompt.txt"), filepath.Join(root, "prompt.txt.improved")},
		{"extensionless", filepath.Join(root, "prompt"), filepath.Join(root, "prompt.improved")},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := promptImprovementOutputPath(root, test.path, options{promptFile: test.path}); got != test.want {
				t.Fatalf("promptImprovementOutputPath() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestInlinePromptNeverUsesDisplayLabelAsSourceFile(t *testing.T) {
	root := t.TempDir()
	labelPath := filepath.Join(root, "CLI --prompt")
	if err := os.WriteFile(labelPath, []byte("unrelated file"), 0644); err != nil {
		t.Fatal(err)
	}
	prompt, source, _, promptFile, err := resolvePrompt(root, "codex", options{prompt: "inline prompt"})
	if err != nil || prompt != "inline prompt" || source != "CLI --prompt" || promptFile != "" {
		t.Fatalf("resolvePrompt() = %q, %q, %q, %v", prompt, source, promptFile, err)
	}
	if got, want := promptImprovementOutputPath(root, promptFile, options{}), filepath.Join(root, ".start-issue", "prompt.improved.md"); got != want {
		t.Fatalf("inline prompt proposal path = %q, want %q", got, want)
	}
}

func TestImprovePromptPassesLegacyRequestToHelper(t *testing.T) {
	bin, root, log := t.TempDir(), t.TempDir(), filepath.Join(t.TempDir(), "pi-request")
	writeExecutable(t, filepath.Join(bin, "pi"), "#!/bin/sh\nprintf '%s' \"$*\" > \"$START_ISSUE_PROMPT_HELPER_LOG\"\nprintf 'improved prompt'\n")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("START_ISSUE_PROMPT_HELPER_LOG", log)

	promptPath := filepath.Join(root, "prompt.md")
	outputPath := filepath.Join(root, "proposal.md")
	err := improvePrompt(root, "pi", "", "Current template", "CLI --prompt-file: "+promptPath, promptPath, options{promptFile: promptPath, promptOutput: outputPath}, issue{Title: "Issue title", Body: "Issue body"}, "owner/repo", "34", "bug, urgent")
	if err != nil {
		t.Fatal(err)
	}
	request, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Prompt source:\nCLI --prompt-file: " + promptPath,
		"Repository:\nowner/repo",
		"Current issue used as improvement context:\n- URL: https://github.com/owner/repo/issues/34\n- Number: 34\n- Title: Issue title\n- Labels: bug, urgent\n- Body:\nIssue body",
		"Current prompt template:\n--- START PROMPT TEMPLATE ---\nCurrent template\n--- END PROMPT TEMPLATE ---",
	} {
		if !strings.Contains(string(request), want) {
			t.Fatalf("helper request missing %q:\n%s", want, request)
		}
	}
}

func TestImprovePromptRejectsEmptyProposal(t *testing.T) {
	bin, root := t.TempDir(), t.TempDir()
	writeExecutable(t, filepath.Join(bin, "pi"), "#!/bin/sh\nprintf '  \\n\\t  \\n'\n")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	outputPath := filepath.Join(root, "proposals", "prompt.md")
	err := improvePrompt(root, "pi", "", "prompt", "built-in default", "", options{promptOutput: outputPath}, issue{Title: "Issue"}, "owner/repo", "1", "")
	if err == nil || !strings.Contains(err.Error(), "proposal is empty") {
		t.Fatalf("improvePrompt() error = %v", err)
	}
	if _, statErr := os.Stat(outputPath); !os.IsNotExist(statErr) {
		t.Fatalf("empty proposal wrote output file: %v", statErr)
	}
}

func TestInstallBinaryRestoresExecutablePermissions(t *testing.T) {
	target := filepath.Join(t.TempDir(), "start-issue")
	if err := os.WriteFile(target, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	previous, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if err := installBinary(target, []byte("new")); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0755 {
		t.Fatalf("installed mode = %o, want 755", got)
	}
	if os.SameFile(previous, info) {
		t.Fatal("installBinary replaced the target in place instead of atomically renaming a staged binary")
	}
	if _, err := os.Stat(target + ".new"); !os.IsNotExist(err) {
		t.Fatalf("staged binary was not removed: %v", err)
	}
}

func TestRunInitWarnsOnFailure(t *testing.T) {
	bin, worktree := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(worktree, "init.sh"), []byte("ignored"), 0644); err != nil {
		t.Fatal(err)
	}
	writeExecutable(t, filepath.Join(bin, "bash"), "#!/bin/sh\nexit 1\n")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	output := captureStdout(t, func() { runInit(worktree, false) })
	if !strings.Contains(output, "Warning: init.sh exited with non-zero code") {
		t.Fatalf("init failure warning missing:\n%s", output)
	}
}

func TestRenameZellijTabWarnsOnFailure(t *testing.T) {
	bin := t.TempDir()
	writeExecutable(t, filepath.Join(bin, "zellij-tab-status"), "#!/bin/sh\nexit 1\n")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	output := captureStdout(t, func() { renameZellijTab("34", false) })
	if !strings.Contains(output, "Warning: Could not rename zellij tab with zellij-tab-status") {
		t.Fatalf("zellij failure warning missing:\n%s", output)
	}
}

func TestRenameZellijTabDryRunReportsMissingOptionalCommand(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	output := captureStdout(t, func() { renameZellijTab("34", true) })
	if !strings.Contains(output, "[DRY-RUN] Would skip zellij tab rename: zellij-tab-status not found") {
		t.Fatalf("missing zellij dry-run message:\n%s", output)
	}
}

func TestUsageListsCompatibilityEntryPoints(t *testing.T) {
	output := captureStdout(t, usage)
	for _, want := range []string{
		"--command, -c <cmd>",
		"--setup",
		"--update",
		"--install",
		"--human-gate-help",
		"Agent selection precedence:",
		".start-issue/agent in the git root",
		"Prompt template precedence:",
		"{ISSUE_URL}, {ISSUE_NUMBER}, {ISSUE_TITLE}, {ISSUE_BODY}, {ISSUE_LABELS}",
		"start-issue https://github.com/owner/repo/issues/123",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("help missing %q:\n%s", want, output)
		}
	}
}

func TestAIBranchPromptPreservesTransliterationAndTagConstraints(t *testing.T) {
	bin, log := t.TempDir(), filepath.Join(t.TempDir(), "prompt")
	writeExecutable(t, filepath.Join(bin, "pi"), "#!/bin/sh\nlast=''\nfor arg do last=$arg; done\nprintf '%s' \"$last\" > '"+log+"'\nprintf '%s\\n' feature/issue-34-ispravit-tsap\n")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	if _, err := aiBranchName("pi", "", t.TempDir(), "34", "[brief] Исправить ЦАП", "bug"); err != nil {
		t.Fatal(err)
	}
	prompt, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"transliterate it to English", "Strip leading bracketed process/stage tags", "[brief]"} {
		if !strings.Contains(string(prompt), want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt)
		}
	}
}

func TestAIBranchNamePreservesCompleteResponseLine(t *testing.T) {
	bin := t.TempDir()
	writeExecutable(t, filepath.Join(bin, "pi"), "#!/bin/sh\nprintf '%s\\n' 'Here is the branch: feature/issue-34-fix-login'\n")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	got, err := aiBranchName("pi", "", t.TempDir(), "34", "Fix login", "bug")
	if err != nil {
		t.Fatal(err)
	}
	if want := "Here is the branch: feature/issue-34-fix-login"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestInstallDryRunDoesNotFetchOrWrite(t *testing.T) {
	home, bin := t.TempDir(), t.TempDir()
	marker := filepath.Join(t.TempDir(), "gh-ran")
	writeExecutable(t, filepath.Join(bin, "gh"), "#!/bin/sh\ntouch '"+marker+"'\n")
	t.Setenv("HOME", home)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	if err := installMode(true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("dry-run fetched the release: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".local", "bin", "start-issue")); !os.IsNotExist(err) {
		t.Fatalf("dry-run installed a binary: %v", err)
	}
}

func TestFirstRunOnboardingReusesBufferedInput(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := maybeRunFirstRunOnboarding(false, "", bufio.NewReader(strings.NewReader("y\n2\ny\n"))); err != nil {
		t.Fatal(err)
	}
	agent, err := os.ReadFile(filepath.Join(home, ".config", "start-issue", "agent"))
	if err != nil || string(agent) != "codex\n" {
		t.Fatalf("agent = %q, %v", agent, err)
	}
}

func TestFirstRunOnboardingSavesClaudeCommandPrompt(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := maybeRunFirstRunOnboarding(false, "/debug", bufio.NewReader(strings.NewReader("y\n1\ny\n"))); err != nil {
		t.Fatal(err)
	}
	prompt, err := os.ReadFile(filepath.Join(home, ".config", "start-issue", "prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(prompt), "/debug {ISSUE_URL}\n"; got != want {
		t.Fatalf("saved prompt = %q, want %q", got, want)
	}
}

func TestFirstRunOnboardingDeclinesWithoutSetup(t *testing.T) {
	for _, response := range []string{"n\n", "no\n"} {
		t.Run(strings.TrimSpace(response), func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)

			if err := maybeRunFirstRunOnboarding(false, "", bufio.NewReader(strings.NewReader(response))); err != nil {
				t.Fatal(err)
			}
			dir := filepath.Join(home, ".config", "start-issue")
			if _, err := os.Stat(dir); err != nil {
				t.Fatalf("first-run marker missing: %v", err)
			}
			for _, name := range []string{"agent", "prompt.md"} {
				if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
					t.Fatalf("declining setup created %s: %v", name, err)
				}
			}
		})
	}
}

func TestRunPerformsFirstRunOnboardingBeforeGitValidation(t *testing.T) {
	home, bin := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", bin)

	var runErr error
	output := captureStdout(t, func() {
		runErr = runWithReader(options{issue: "1"}, bufio.NewReader(strings.NewReader("n\n")))
	})
	if runErr == nil || runErr.Error() != "git not found" {
		t.Fatalf("run error = %v, want git validation failure", runErr)
	}
	if !strings.Contains(output, "Configuration is not initialized yet.") || !strings.Contains(output, "Run setup now? [Y/n]") {
		t.Fatalf("first-run onboarding was not offered before git validation:\n%s", output)
	}
	if _, err := os.Stat(filepath.Join(home, ".config", "start-issue")); err != nil {
		t.Fatalf("first-run marker missing: %v", err)
	}
}

func TestFirstRunOnboardingRejectsEOFAndInvalidResponse(t *testing.T) {
	for _, response := range []string{"", "maybe\n"} {
		t.Run(fmt.Sprintf("%q", response), func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			err := maybeRunFirstRunOnboarding(false, "", bufio.NewReader(strings.NewReader(response)))
			if err == nil {
				t.Fatal("expected response error")
			}
			if _, statErr := os.Stat(filepath.Join(home, ".config", "start-issue")); !os.IsNotExist(statErr) {
				t.Fatalf("invalid response initialized config: %v", statErr)
			}
		})
	}
}

func TestSetupRejectsEOFAndHonorsNo(t *testing.T) {
	for _, test := range []struct {
		name, input string
		wantErr     bool
	}{
		{name: "eof", input: "2\n", wantErr: true},
		{name: "no", input: "2\nno\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			home := t.TempDir()
			err := setupMode(home, false, "", bufio.NewReader(strings.NewReader(test.input)))
			if (err != nil) != test.wantErr {
				t.Fatalf("setup error = %v, want error: %t", err, test.wantErr)
			}
			prompt := filepath.Join(home, ".config", "start-issue", "prompt.md")
			if _, err := os.Stat(prompt); !os.IsNotExist(err) {
				t.Fatalf("setup wrote prompt after %s: %v", test.name, err)
			}
			if _, err := os.Stat(filepath.Join(home, ".config", "start-issue")); err != nil {
				t.Fatalf("setup did not create configuration marker after %s: %v", test.name, err)
			}
		})
	}
}

func TestSetupReturnsConfigurationRemovalErrors(t *testing.T) {
	for _, test := range []struct {
		name, input, config, message string
	}{
		{name: "prompt", input: "codex\nno\n", config: "prompt.md", message: "remove prompt template"},
		{name: "agent", input: "skip\nno\n", config: "agent", message: "remove agent config"},
	} {
		t.Run(test.name, func(t *testing.T) {
			home := t.TempDir()
			path := filepath.Join(home, ".config", "start-issue", test.config)
			if err := os.MkdirAll(path, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(path, "keep"), []byte("keep"), 0644); err != nil {
				t.Fatal(err)
			}

			err := setupMode(home, false, "", bufio.NewReader(strings.NewReader(test.input)))
			if err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("setup error = %v", err)
			}
		})
	}
}

func TestSetupPreviewsSelectedPromptBeforeConfirmation(t *testing.T) {
	home := t.TempDir()
	output := captureStdout(t, func() {
		if err := setupMode(home, false, "", bufio.NewReader(strings.NewReader("2\ny\n"))); err != nil {
			t.Fatal(err)
		}
	})
	preview := "Default prompt preview:\n" + defaultPortablePrompt()
	if !strings.Contains(output, preview) {
		t.Fatalf("setup output did not preview the selected prompt:\n%s", output)
	}
	if strings.Index(output, "Default prompt preview:") > strings.Index(output, "Save a default prompt?") {
		t.Fatalf("setup asked for confirmation before previewing the prompt:\n%s", output)
	}
}

func TestSetupSavesClaudeCommandPrompt(t *testing.T) {
	home := t.TempDir()
	if err := setupMode(home, false, "/debug", bufio.NewReader(strings.NewReader("claude\ny\n"))); err != nil {
		t.Fatal(err)
	}
	prompt, err := os.ReadFile(filepath.Join(home, ".config", "start-issue", "prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(prompt), "/debug {ISSUE_URL}\n"; got != want {
		t.Fatalf("saved prompt = %q, want %q", got, want)
	}
}

func TestSetupAcceptsLegacyAgentSelections(t *testing.T) {
	for _, test := range []struct {
		choice    string
		wantAgent string
	}{
		{choice: "claude", wantAgent: "claude"},
		{choice: "Claude", wantAgent: "claude"},
		{choice: "codex", wantAgent: "codex"},
		{choice: "Codex", wantAgent: "codex"},
		{choice: "kimi", wantAgent: "kimi"},
		{choice: "Kimi", wantAgent: "kimi"},
		{choice: "pi", wantAgent: "pi"},
		{choice: "Pi", wantAgent: "pi"},
		{choice: "skip", wantAgent: ""},
		{choice: "Skip", wantAgent: ""},
		{choice: "", wantAgent: ""},
	} {
		t.Run(test.choice, func(t *testing.T) {
			home := t.TempDir()
			if err := setupMode(home, false, "", bufio.NewReader(strings.NewReader(test.choice+"\nno\n"))); err != nil {
				t.Fatal(err)
			}
			agent, err := os.ReadFile(filepath.Join(home, ".config", "start-issue", "agent"))
			if test.wantAgent == "" {
				if !os.IsNotExist(err) {
					t.Fatalf("setup agent config = %q, %v; want no agent config", agent, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got := string(agent); got != test.wantAgent+"\n" {
				t.Fatalf("setup agent config = %q, want %q", got, test.wantAgent+"\n")
			}
		})
	}
}

func TestSetupDryRunPrintsInteractiveConfigPlanWithoutWriting(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "start-issue")
	output := captureStdout(t, func() {
		if err := setupMode(home, true, "", bufio.NewReader(strings.NewReader("5\nn\n"))); err != nil {
			t.Fatal(err)
		}
	})
	for _, want := range []string{
		"Default prompt preview:\n/task-router:route-task {ISSUE_URL}",
		"Would create configuration in: " + dir,
		"Would remove prompt template: " + filepath.Join(dir, "prompt.md"),
		"Would remove agent config: " + filepath.Join(dir, "agent"),
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("dry-run output missing %q:\n%s", want, output)
		}
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("setup dry-run created configuration: %v", err)
	}
}

func TestEmptyRootDoesNotReadRelativeProjectConfig(t *testing.T) {
	wd := t.TempDir()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(previous) }()
	if err := os.Chdir(wd); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(".start-issue", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(".start-issue", "agent"), []byte("codex\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", t.TempDir())

	agent, source, err := resolveAgent("", "")
	if err != nil || agent != "claude" || source != "built-in default" {
		t.Fatalf("got %q %q %v", agent, source, err)
	}
}

func TestResolvePromptPrefersProjectConfig(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".start-issue")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "prompt.md"), []byte("project {ISSUE_URL}"), 0644); err != nil {
		t.Fatal(err)
	}
	got, source, location, promptFile, err := resolvePrompt(root, "codex", options{})
	if err != nil || got != "project {ISSUE_URL}" || source != filepath.Join(dir, "prompt.md") || location != filepath.Join(dir, "prompt.md") {
		t.Fatalf("got %q %q %q %v", got, source, location, err)
	}
	if promptFile != filepath.Join(dir, "prompt.md") {
		t.Fatalf("prompt file = %q, want project prompt path", promptFile)
	}
}

func TestResolvePromptTracksSourceAndLocationSeparately(t *testing.T) {
	inline, source, location, inlinePromptFile, err := resolvePrompt(t.TempDir(), "codex", options{prompt: "hello"})
	if err != nil || inline != "hello" || source != "CLI --prompt" || location != "inline CLI argument" {
		t.Fatalf("inline = %q %q %q %v", inline, source, location, err)
	}
	if inlinePromptFile != "" {
		t.Fatalf("inline prompt file = %q, want empty", inlinePromptFile)
	}

	promptFile := filepath.Join(t.TempDir(), "prompt with spaces.md")
	if err := os.WriteFile(promptFile, []byte("from file"), 0644); err != nil {
		t.Fatal(err)
	}
	_, source, location, promptFile, err = resolvePrompt(t.TempDir(), "codex", options{promptFile: promptFile})
	if err != nil || source != "CLI --prompt-file: "+promptFile || location != promptFile {
		t.Fatalf("file = %q %q %v", source, location, err)
	}
	if promptFile == "" {
		t.Fatal("file-backed prompt did not retain its source path")
	}
}

func TestInitScopeDoesNotImportOtherScopeAgent(t *testing.T) {
	root, home, bin := t.TempDir(), t.TempDir(), t.TempDir()
	writeExecutable(t, filepath.Join(bin, "git"), "#!/bin/sh\nprintf '%s\\n' '"+root+"'\n")
	t.Setenv("HOME", home)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	if err := os.MkdirAll(filepath.Join(home, ".config", "start-issue"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".config", "start-issue", "agent"), []byte("codex\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := runMode(options{mode: "init", project: true}); err != nil {
		t.Fatal(err)
	}
	agent, err := os.ReadFile(filepath.Join(root, ".start-issue", "agent"))
	if err != nil || string(agent) != "claude\n" {
		t.Fatalf("project agent = %q, %v", agent, err)
	}

	if err := os.RemoveAll(filepath.Join(home, ".config", "start-issue")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".start-issue", "agent"), []byte("codex\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := runMode(options{mode: "init", user: true}); err != nil {
		t.Fatal(err)
	}
	agent, err = os.ReadFile(filepath.Join(home, ".config", "start-issue", "agent"))
	if err != nil || string(agent) != "claude\n" {
		t.Fatalf("user agent = %q, %v", agent, err)
	}
}

func TestInitDryRunValidatesAndPrintsPlanWithoutWriting(t *testing.T) {
	root, home, bin := t.TempDir(), t.TempDir(), t.TempDir()
	writeExecutable(t, filepath.Join(bin, "git"), "#!/bin/sh\nprintf '%s\\n' '"+root+"'\n")
	t.Setenv("HOME", home)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	err := runMode(options{mode: "init", project: true, dryRun: true, agent: "unknown"})
	if err == nil || !strings.Contains(err.Error(), "Unknown agent") {
		t.Fatalf("unknown agent error = %v", err)
	}
	err = runMode(options{mode: "init", project: true, dryRun: true, promptFile: filepath.Join(root, "missing.md")})
	if err == nil || !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("missing prompt error = %v", err)
	}

	output := captureStdout(t, func() {
		err = runMode(options{mode: "init", project: true, dryRun: true, agent: "codex"})
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Would create configuration in: " + filepath.Join(root, ".start-issue"),
		"Would write agent config: " + filepath.Join(root, ".start-issue", "agent"),
		"Would write prompt template: " + filepath.Join(root, ".start-issue", "prompt.md"),
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("dry-run output missing %q:\n%s", want, output)
		}
	}
	if _, statErr := os.Stat(filepath.Join(root, ".start-issue")); !os.IsNotExist(statErr) {
		t.Fatalf("dry-run created config: %v", statErr)
	}
}

func TestInitDefaultPromptUsesAgentPersistedAtTarget(t *testing.T) {
	root, home, bin := t.TempDir(), t.TempDir(), t.TempDir()
	configDir := filepath.Join(root, ".start-issue")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "agent"), []byte("codex\n"), 0644); err != nil {
		t.Fatal(err)
	}
	writeExecutable(t, filepath.Join(bin, "git"), "#!/bin/sh\nprintf '%s\\n' '"+root+"'\n")
	t.Setenv("HOME", home)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	if err := runMode(options{mode: "init", project: true, agent: "claude"}); err != nil {
		t.Fatal(err)
	}
	prompt, err := os.ReadFile(filepath.Join(configDir, "prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(prompt), "/task-router:route-task") {
		t.Fatalf("Codex config received Claude prompt: %q", prompt)
	}
	if !strings.Contains(string(prompt), "Implement GitHub issue {ISSUE_URL}") {
		t.Fatalf("got unexpected prompt: %q", prompt)
	}
}

func TestInitDefaultClaudePromptUsesCommand(t *testing.T) {
	root, home, bin := t.TempDir(), t.TempDir(), t.TempDir()
	writeExecutable(t, filepath.Join(bin, "git"), "#!/bin/sh\nprintf '%s\\n' '"+root+"'\n")
	t.Setenv("HOME", home)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	if err := runMode(options{mode: "init", project: true, command: "/custom:route"}); err != nil {
		t.Fatal(err)
	}
	prompt, err := os.ReadFile(filepath.Join(root, ".start-issue", "prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(prompt), "/custom:route {ISSUE_URL}\n"; got != want {
		t.Fatalf("prompt = %q, want %q", got, want)
	}
}

func TestInitPromptFileNormalizesTrailingNewlines(t *testing.T) {
	root, home, bin, source := t.TempDir(), t.TempDir(), t.TempDir(), filepath.Join(t.TempDir(), "source.md")
	writeExecutable(t, filepath.Join(bin, "git"), "#!/bin/sh\nprintf '%s\\n' '"+root+"'\n")
	if err := os.WriteFile(source, []byte("Prompt {ISSUE_URL}\n\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	if err := runMode(options{mode: "init", project: true, promptFile: source}); err != nil {
		t.Fatal(err)
	}
	prompt, err := os.ReadFile(filepath.Join(root, ".start-issue", "prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(prompt), "Prompt {ISSUE_URL}\n"; got != want {
		t.Fatalf("prompt = %q, want %q", got, want)
	}
}

func TestLaunchArgsNoneIsEmpty(t *testing.T) {
	if got := launchArgs("none", "", "", ""); len(got) != 0 {
		t.Fatalf("got %q", got)
	}
}

func TestPrintLaunchShellQuotesArguments(t *testing.T) {
	output := captureStdout(t, func() {
		printLaunch("codex", "", "/tmp/work tree", "# prompt; touch not-run")
	})
	if !strings.Contains(output, "codex --cd '/tmp/work tree' --dangerously-bypass-approvals-and-sandbox '# prompt; touch not-run'") {
		t.Fatalf("launch output is not shell quoted:\n%s", output)
	}
	if got := shellJoin([]string{"", "~", "#prompt"}); got != "'' '~' '#prompt'" {
		t.Fatalf("shellJoin = %q", got)
	}
}

func TestPrintLaunchLargePromptDiagnostics(t *testing.T) {
	prompt := strings.Repeat("x", 4001)
	output := captureStdout(t, func() {
		printLaunch("codex", "", "/tmp/work tree", prompt)
	})
	for _, want := range []string{
		"Prompt length: 4001 chars",
		"Prompt omitted from command display because it is large.",
		"Set START_ISSUE_DUMP_PROMPT=1 to print the full rendered prompt.",
		"<rendered prompt: 4001 chars>",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("dry-run diagnostics missing %q:\n%s", want, output)
		}
	}
	if strings.Contains(output, prompt) {
		t.Fatalf("large rendered prompt was not omitted from command display:\n%s", output)
	}
}

func TestPrintLaunchCountsUnicodePromptCharacters(t *testing.T) {
	prompt := strings.Repeat("я", 3000)
	output := captureStdout(t, func() {
		printLaunch("codex", "", "/tmp/work tree", prompt)
	})
	if !strings.Contains(output, "Prompt length: 3000 chars") {
		t.Fatalf("unicode prompt character count is wrong:\n%s", output)
	}
	if strings.Contains(output, "Prompt omitted from command display because it is large.") {
		t.Fatalf("unicode prompt below the character threshold was omitted:\n%s", output)
	}
}

func TestPrintLaunchShowsWorktreeCWDForClaudeAndPi(t *testing.T) {
	for _, agent := range []string{"claude", "pi"} {
		t.Run(agent, func(t *testing.T) {
			output := captureStdout(t, func() {
				printLaunch(agent, "", "/tmp/work tree", "prompt")
			})
			if !strings.Contains(output, "Would run: cd '/tmp/work tree' && "+agent) {
				t.Fatalf("dry-run launch plan omits worktree cwd:\n%s", output)
			}
		})
	}
}

func TestLaunchSelectedReportsAgentHandoff(t *testing.T) {
	bin := t.TempDir()
	writeExecutable(t, filepath.Join(bin, "codex"), "#!/bin/sh\nexit 0\n")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	worktree := t.TempDir()
	output := captureStdout(t, func() {
		if err := launchSelected(options{}, "codex", "", worktree, "prompt"); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(output, "Handing off to codex in "+worktree) {
		t.Fatalf("agent handoff was not reported:\n%s", output)
	}
}

func TestCanonicalPathMakesRelativeWorktreeAbsolute(t *testing.T) {
	if got := canonicalPath(filepath.Join("worktrees", "feature", "issue-34")); !filepath.IsAbs(got) {
		t.Fatalf("worktree path is relative: %q", got)
	}
}

func TestLaunchPreservesAgentExitCode(t *testing.T) {
	bin := t.TempDir()
	writeExecutable(t, filepath.Join(bin, "codex"), "#!/bin/sh\nexit 42\n")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	err := launch("codex", "", t.TempDir(), "prompt")
	var exit exitError
	if !errors.As(err, &exit) || exit.code != 42 {
		t.Fatalf("got %T %v", err, err)
	}
}

func TestLaunchPreservesSignalDerivedExitCode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX signal exit statuses are not available on Windows")
	}
	bin := t.TempDir()
	writeExecutable(t, filepath.Join(bin, "codex"), "#!/bin/sh\nkill -INT $$\n")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	err := launch("codex", "", t.TempDir(), "prompt")
	var exit exitError
	if !errors.As(err, &exit) || exit.code != 130 {
		t.Fatalf("got %T %v", err, err)
	}
}

func TestLaunchUsesAdapterSpecificWorkingDirectory(t *testing.T) {
	bin, worktree, log := t.TempDir(), t.TempDir(), filepath.Join(t.TempDir(), "cwd")
	for _, agent := range []string{"claude", "codex", "kimi", "pi"} {
		writeExecutable(t, filepath.Join(bin, agent), "#!/bin/sh\npwd > \"$START_ISSUE_CWD_LOG\"\n")
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("START_ISSUE_CWD_LOG", log)
	caller, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for _, agent := range []string{"claude", "codex", "kimi", "pi"} {
		t.Run(agent, func(t *testing.T) {
			if err := launch(agent, "", worktree, "prompt"); err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(log)
			if err != nil {
				t.Fatal(err)
			}
			want := caller
			if agent == "claude" || agent == "kimi" || agent == "pi" {
				want = worktree
			}
			if strings.TrimSpace(string(got)) != want {
				t.Fatalf("working directory = %q, want %q", strings.TrimSpace(string(got)), want)
			}
		})
	}
}

func TestHelperArgsAreNonInteractive(t *testing.T) {
	pi := helperArgs("pi", "", "/repo", "prompt")
	if got := fmt.Sprint(pi); got != "[pi --print --no-tools --no-session prompt]" {
		t.Fatalf("pi helper args: %s", got)
	}
	kimi := helperArgs("kimi", "model", "/repo", "prompt")
	if got := fmt.Sprint(kimi); got != "[kimi --model model -p prompt]" {
		t.Fatalf("kimi helper args: %s", got)
	}
}

func TestHumanGateSavesThreadIDBeforeDone(t *testing.T) {
	worktree, bin := t.TempDir(), t.TempDir()
	writeFakeCodex(t, bin)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("START_ISSUE_RUN_ID", "done")
	t.Setenv("CODEX_EVENTS", `{"type":"thread.started","thread_id":"thread-done"}`)
	t.Setenv("CODEX_LAST", "STATUS: DONE")
	t.Setenv("START_ISSUE_FAKE_CODEX_REJECT_ASK_FOR_APPROVAL", "1")

	if err := humanGate("", worktree, "prompt", false); err != nil {
		t.Fatal(err)
	}
	threadID, err := os.ReadFile(filepath.Join(worktree, ".start-issue", "runs", "done", "thread-id"))
	if err != nil || string(threadID) != "thread-done\n" {
		t.Fatalf("thread-id = %q, %v", threadID, err)
	}
}

func TestHumanGateSavesThreadIDWhenFinalMessageIsMissing(t *testing.T) {
	worktree, bin := t.TempDir(), t.TempDir()
	writeFakeCodex(t, bin)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("START_ISSUE_RUN_ID", "missing-last-message")
	t.Setenv("CODEX_EVENTS", `{"type":"thread.started","thread_id":"thread-recovery"}`)
	t.Setenv("CODEX_SKIP_LAST", "1")

	err := humanGate("", worktree, "prompt", false)
	if err == nil || !strings.Contains(err.Error(), "No recognized final status found") {
		t.Fatalf("humanGate error = %v, want missing final-status error", err)
	}
	threadID, readErr := os.ReadFile(filepath.Join(worktree, ".start-issue", "runs", "missing-last-message", "thread-id"))
	if readErr != nil || string(threadID) != "thread-recovery\n" {
		t.Fatalf("thread-id = %q, %v", threadID, readErr)
	}
}

func TestHumanGateExecFailureReturnsExitCodeOne(t *testing.T) {
	worktree, bin := t.TempDir(), t.TempDir()
	writeFakeCodex(t, bin)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("START_ISSUE_RUN_ID", "exec-failure")
	t.Setenv("CODEX_EVENTS", `{"type":"thread.started","thread_id":"thread-failure"}`)
	t.Setenv("CODEX_LAST", "STATUS: DONE")
	t.Setenv("CODEX_EXEC_EXIT", "42")

	err := humanGate("", worktree, "prompt", false)
	var exit exitError
	if !errors.As(err, &exit) || exit.code != 1 {
		t.Fatalf("got %T %v, want human-gate exit code 1", err, err)
	}
	if !strings.Contains(err.Error(), "Codex batch run failed") {
		t.Fatalf("error = %v, want batch failure diagnostic", err)
	}
	threadID, readErr := os.ReadFile(filepath.Join(worktree, ".start-issue", "runs", "exec-failure", "thread-id"))
	if readErr != nil || string(threadID) != "thread-failure\n" {
		t.Fatalf("thread-id = %q, %v", threadID, readErr)
	}
}

func TestRunChecksForGitBeforeRepositoryValidation(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	err := runWithReader(options{}, bufio.NewReader(strings.NewReader("")))
	if err == nil || err.Error() != "git not found" {
		t.Fatalf("runWithReader error = %v, want git dependency error", err)
	}
}

func TestHumanGateDryRunShowsAllStateArtifacts(t *testing.T) {
	worktree := t.TempDir()
	t.Setenv("START_ISSUE_RUN_ID", "plan")
	dir := filepath.Join(worktree, ".start-issue", "runs", "plan")
	output := captureStdout(t, func() {
		if err := humanGate("", worktree, "prompt", true); err != nil {
			t.Fatal(err)
		}
	})
	for _, want := range []string{
		"--output-last-message " + filepath.Join(dir, "last-message.txt"),
		"> " + filepath.Join(dir, "events.jsonl"),
		"Would write captured thread ID: " + filepath.Join(dir, "thread-id"),
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("dry-run output missing %q:\n%s", want, output)
		}
	}
	if strings.Contains(output, "--ask-for-approval") {
		t.Fatalf("dry-run includes obsolete --ask-for-approval argument:\n%s", output)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("human-gate dry-run created state directory: %v", err)
	}
}

func TestHumanGatePreservesCallerWorkingDirectory(t *testing.T) {
	worktree, bin, log := t.TempDir(), t.TempDir(), filepath.Join(t.TempDir(), "cwd")
	writeFakeCodex(t, bin)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("START_ISSUE_CWD_LOG", log)
	t.Setenv("START_ISSUE_RUN_ID", "cwd")
	t.Setenv("CODEX_EVENTS", `{"type":"thread.started","thread_id":"thread-cwd"}`)
	t.Setenv("CODEX_LAST", "STATUS: DONE")
	want, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := humanGate("", worktree, "prompt", false); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(got)) != want {
		t.Fatalf("working directory = %q, want %q", strings.TrimSpace(string(got)), want)
	}
}

func TestHumanGateRejectsDoneWithoutThreadID(t *testing.T) {
	worktree, bin := t.TempDir(), t.TempDir()
	writeFakeCodex(t, bin)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("START_ISSUE_RUN_ID", "missing-thread")
	t.Setenv("CODEX_EVENTS", `{"type":"item.completed"}`)
	t.Setenv("CODEX_LAST", "STATUS: DONE")

	err := humanGate("", worktree, "prompt", false)
	if err == nil || !strings.Contains(err.Error(), "did not capture thread_id") {
		t.Fatalf("got %v", err)
	}
}

func TestHumanGateResumeFailureReturnsExitCodeTwo(t *testing.T) {
	worktree, bin := t.TempDir(), t.TempDir()
	writeFakeCodex(t, bin)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("START_ISSUE_RUN_ID", "resume-failure")
	t.Setenv("CODEX_EVENTS", `{"type":"thread.started","thread_id":"thread-resume"}`)
	t.Setenv("CODEX_LAST", "STATUS: HUMAN_GATE")
	t.Setenv("CODEX_RESUME_EXIT", "1")

	err := humanGate("", worktree, "prompt", false)
	var exit exitError
	if !errors.As(err, &exit) || exit.code != 2 {
		t.Fatalf("got %T %v", err, err)
	}
}

func writeExecutable(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}
}

func writeFakeCodex(t *testing.T, bin string) {
	t.Helper()
	writeExecutable(t, filepath.Join(bin, "codex"), `#!/bin/sh
if [ "$START_ISSUE_FAKE_CODEX_REJECT_ASK_FOR_APPROVAL" = "1" ] && [ "${*#*--ask-for-approval}" != "$*" ]; then
  printf '%s\n' "unexpected obsolete --ask-for-approval flag" >&2
  exit 1
fi
if [ -n "$START_ISSUE_CWD_LOG" ]; then
  pwd > "$START_ISSUE_CWD_LOG"
fi
if [ "$1" = "exec" ]; then
  last=""
  while [ "$#" -gt 0 ]; do
    if [ "$1" = "--output-last-message" ]; then
      last="$2"
      shift 2
      continue
    fi
    shift
  done
  printf '%s\n' "$CODEX_EVENTS"
  if [ "$CODEX_SKIP_LAST" != "1" ]; then
    printf '%s\n' "$CODEX_LAST" > "$last"
  fi
  exit "${CODEX_EXEC_EXIT:-0}"
fi
if [ "$1" = "resume" ]; then
  exit "${CODEX_RESUME_EXIT:-0}"
fi
exit 1
`)
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	previous := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = writer
	defer func() { os.Stdout = previous }()
	fn()
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	return string(output)
}
