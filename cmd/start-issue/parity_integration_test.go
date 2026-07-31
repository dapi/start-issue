package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// TestCLIParityHelper runs the public command in a subprocess so the parity
// cases exercise argument parsing, external command fakes, and side effects.
func TestCLIParityHelper(t *testing.T) {
	if os.Getenv("START_ISSUE_PARITY_HELPER") != "1" {
		return
	}
	for index, arg := range os.Args {
		if arg == "--" {
			os.Args = append([]string{"start-issue"}, os.Args[index+1:]...)
			main()
			return
		}
	}
	t.Fatal("missing CLI argument separator")
}

func TestCLIParityIssueFetchConfigWorktreeAndReuse(t *testing.T) {
	fixture := newParityFixture(t)

	projectConfig := filepath.Join(fixture.repo, ".start-issue")
	if err := os.MkdirAll(projectConfig, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectConfig, "agent"), []byte("none\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixture.home, ".config", "start-issue", "agent"), []byte("codex\n"), 0644); err != nil {
		t.Fatal(err)
	}

	output, err := fixture.run("", "1", "--dry-run", "--no-init")
	if err != nil {
		t.Fatalf("dry-run failed: %v\n%s", err, output)
	}
	for _, want := range []string{
		"Agent: none",
		"Agent source: " + filepath.Join(projectConfig, "agent"),
		"Fetching issue #1 from owner/repo",
		"Would run: git worktree add -b feature/issue-1-add-login-button",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("dry-run output missing %q:\n%s", want, output)
		}
	}
	if log, err := os.ReadFile(fixture.ghLog); err != nil || !strings.Contains(string(log), "auth status") {
		t.Fatalf("gh authentication preflight not recorded: %q, %v", log, err)
	}

	output, err = fixture.run("", "1", "--agent", "none", "--no-init")
	if err != nil {
		t.Fatalf("creation failed: %v\n%s", err, output)
	}
	if log, err := os.ReadFile(fixture.gitLog); err != nil || !strings.Contains(string(log), "worktree add -b feature/issue-1-add-login-button") {
		t.Fatalf("worktree creation not recorded: %q, %v", log, err)
	}

	worktree := canonicalPath(filepath.Join(fixture.worktrees, "feature", "issue-1-add-login-button"))
	if err := os.MkdirAll(worktree, 0755); err != nil {
		t.Fatal(err)
	}
	fixture.pathBranch = true
	output, err = fixture.run("1\n", "1", "--agent", "none", "--no-init", "--worktree-dir", fixture.worktrees)
	if err != nil {
		t.Fatalf("reuse failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "Worktree ready at: "+worktree) {
		t.Fatalf("reuse output does not identify existing worktree:\n%s", output)
	}
}

func TestCLIPrintsVersionBannerOnExecutionPaths(t *testing.T) {
	for _, test := range []struct {
		name  string
		input string
		args  []string
	}{
		{name: "issue workflow", args: []string{"1", "--agent", "none", "--no-init", "--dry-run"}},
		{name: "init", args: []string{"init", "--project", "--force", "--agent", "none"}},
		{name: "setup", input: "\n\n", args: []string{"setup", "--dry-run"}},
		// The fixture's release response is intentionally incomplete, but the
		// banner must precede any update diagnostic.
		{name: "update", args: []string{"update", "--dry-run"}},
		{name: "human-gate help", args: []string{"--human-gate-help"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			output, _ := newParityFixture(t).run(test.input, test.args...)
			if !strings.HasPrefix(output, "start-issue v") {
				t.Fatalf("output does not start with the version banner:\n%s", output)
			}
		})
	}
}

func TestCLIReusesOneReaderForOnboardingAndConflictResolution(t *testing.T) {
	fixture := newParityFixture(t)
	if err := os.RemoveAll(filepath.Join(fixture.home, ".config", "start-issue")); err != nil {
		t.Fatal(err)
	}
	fixture.branchExists = true
	if err := os.MkdirAll(fixture.fakeWorktree(), 0755); err != nil {
		t.Fatal(err)
	}

	output, err := fixture.run("n\n1\n", "1", "--agent", "none", "--no-init")
	if err != nil {
		t.Fatalf("onboarding then reuse failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "Worktree ready at: "+fixture.fakeWorktree()) {
		t.Fatalf("conflict response was not available after onboarding:\n%s", output)
	}
}

func TestCLIDiagnosticsReportWorktreeDirectorySource(t *testing.T) {
	fixture := newParityFixture(t)

	output, err := fixture.run("", "1", "--agent", "none", "--no-init", "--dry-run")
	if err != nil {
		t.Fatalf("environment-source dry-run failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "Worktree directory: "+fixture.worktrees+" (START_ISSUE_WORKTREE_DIR)") {
		t.Fatalf("environment worktree source missing from diagnostics:\n%s", output)
	}

	output, err = fixture.run("", "1", "--agent", "none", "--no-init", "--dry-run", "--worktree-dir", fixture.worktrees)
	if err != nil {
		t.Fatalf("CLI-source dry-run failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "Worktree directory: "+fixture.worktrees+" (CLI)") {
		t.Fatalf("CLI worktree source missing from diagnostics:\n%s", output)
	}
}

func TestCLIDiagnosticsReportPromptSourceAndLocation(t *testing.T) {
	fixture := newParityFixture(t)
	prompt := filepath.Join(t.TempDir(), "prompt.md")
	if err := os.WriteFile(prompt, []byte("Implement {ISSUE_URL}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	output, err := fixture.run("", "1", "--agent", "none", "--no-init", "--dry-run", "--prompt-file", prompt)
	if err != nil {
		t.Fatalf("prompt-file dry-run failed: %v\n%s", err, output)
	}
	for _, want := range []string{"Prompt source: CLI --prompt-file: " + prompt, "Prompt location: " + prompt} {
		if !strings.Contains(output, want) {
			t.Fatalf("prompt diagnostics missing %q:\n%s", want, output)
		}
	}
}

func TestCLIParityInitHookAndNoInit(t *testing.T) {
	baseline := extractBashParityBaseline(t)
	for _, test := range []struct {
		name string
		args []string
		want bool
	}{
		{name: "runs init hook", args: []string{"1", "--agent", "none"}, want: true},
		{name: "skips init hook", args: []string{"1", "--agent", "none", "--no-init"}, want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			bashFixture, goFixture := newParityFixture(t), newParityFixture(t)
			bashFixture.initMarker = filepath.Join(t.TempDir(), "bash-init-ran")
			goFixture.initMarker = filepath.Join(t.TempDir(), "go-init-ran")

			want := bashFixture.runBaseline(t, baseline, "", test.args...)
			got := goFixture.runResult("", test.args...)
			if want.exitCode != got.exitCode {
				t.Fatalf("init case exit code differs: Bash=%d Go=%d", want.exitCode, got.exitCode)
			}
			for _, marker := range []string{bashFixture.initMarker, goFixture.initMarker} {
				_, err := os.Stat(marker)
				if (err == nil) != test.want {
					t.Fatalf("init marker %s present=%t, want %t (stat error: %v)", marker, err == nil, test.want, err)
				}
			}
		})
	}
}

func TestCLIParityPreservesCaseSensitiveLabelMatching(t *testing.T) {
	baseline := extractBashParityBaseline(t)
	bashFixture, goFixture := newParityFixture(t), newParityFixture(t)
	bashFixture.issueLabel = "Bug"
	goFixture.issueLabel = "Bug"
	args := []string{"1", "--agent", "none", "--no-init", "--dry-run"}

	want := bashFixture.runBaseline(t, baseline, "", args...)
	got := goFixture.runResult("", args...)
	assertParityResult(t, want, got, []parityOutputRecord{
		lineRecord("labels", `Labels: Bug`),
		branchRecord(),
	})
}

func TestCLIParityLargePromptDiagnostics(t *testing.T) {
	baseline := extractBashParityBaseline(t)
	bashFixture, goFixture := newParityFixture(t), newParityFixture(t)
	prompt := strings.Repeat("x", 4001)
	args := []string{"1", "--agent", "codex", "--no-init", "--dry-run", "--prompt", prompt}

	want := bashFixture.runBaseline(t, baseline, "", args...)
	got := goFixture.runResult("", args...)
	assertParityResult(t, want, got, []parityOutputRecord{
		lineRecord("prompt length", `Prompt length: 4001 chars`),
		lineRecord("prompt omission", `Prompt omitted from command display because it is large\.`),
		lineRecord("prompt omission guidance", `Set START_ISSUE_DUMP_PROMPT=1 to print the full rendered prompt\.`),
		shellEscapedLineRecord("abbreviated rendered prompt", `\\?<rendered\\? prompt:\\? 4001\\? chars\\?>`),
	})
}

func TestParityNormalizationPreservesSemanticRootIdentity(t *testing.T) {
	fixture := newParityFixture(t)
	// Worktree parents are allowed to be nested under a repository. The
	// normalizer must retain that semantic ownership instead of reducing both
	// paths to an indistinguishable temporary-directory token.
	fixture.worktrees = filepath.Join(fixture.repo, "worktrees")
	if err := os.MkdirAll(fixture.worktrees, 0755); err != nil {
		t.Fatal(err)
	}
	for path := range map[string]struct{}{
		filepath.Join(fixture.home, "home-only"):          {},
		filepath.Join(fixture.repo, "repo-only"):          {},
		filepath.Join(fixture.worktrees, "worktree-only"): {},
	} {
		if err := os.WriteFile(path, nil, 0644); err != nil {
			t.Fatal(err)
		}
	}

	normalized := normalizeRawParityOutput(strings.Join([]string{
		filepath.Join(fixture.home, "home-only"),
		filepath.Join(fixture.repo, "repo-only"),
		filepath.Join(fixture.worktrees, "worktree-only"),
	}, "\n"), fixture)
	for _, want := range []string{
		"<HOME>/home-only",
		"<REPO>/repo-only",
		"<WORKTREES>/worktree-only",
	} {
		if !strings.Contains(normalized, want) {
			t.Fatalf("normalized output missing %q:\n%s", want, normalized)
		}
	}

	filesystem := strings.Join(parityFilesystem(fixture), "\n")
	for _, want := range []string{
		"<HOME>/home-only",
		"<REPO>/repo-only",
		"<WORKTREES>/worktree-only",
	} {
		if !strings.Contains(filesystem, want) {
			t.Fatalf("filesystem record missing %q:\n%s", want, filesystem)
		}
	}
}

func TestBaselineAndGoParityForCriticalIssueWorkflows(t *testing.T) {
	baseline := extractBashParityBaseline(t)
	for _, test := range []struct {
		name                    string
		args                    []string
		input                   string
		setup                   func(t *testing.T, fixture *parityFixture)
		strict                  bool
		intentionalDifferenceID string
		records                 []parityOutputRecord
		assertOutcome           func(t *testing.T, result parityResult)
		assertBaseline          func(t *testing.T, result parityResult)
		assertGo                func(t *testing.T, result parityResult)
	}{
		{
			name:   "help",
			args:   []string{"--help"},
			strict: false,
			assertOutcome: func(t *testing.T, result parityResult) {
				assertParityOutputContains(t, result, "Usage: start-issue <issue-url-or-number> [options]", "--agent")
			},
		},
		{
			name:    "invalid-input",
			args:    []string{"--not-an-option"},
			strict:  true,
			records: []parityOutputRecord{lineRecord("unknown-option diagnostic", `Unknown option: --not-an-option`)},
			assertOutcome: func(t *testing.T, result parityResult) {
				if result.exitCode == 0 {
					t.Fatal("invalid option unexpectedly succeeded")
				}
				assertParityOutputContains(t, result, "Unknown option: --not-an-option")
			},
		},
		{
			name: "configuration-precedence",
			args: []string{"1", "--no-init", "--dry-run"},
			records: []parityOutputRecord{
				lineRecord("agent", `Agent: .+`), lineRecord("agent source", `Agent source: .+`),
				lineRecord("model", `Model: .+`), lineRecord("model source", `Model source: .+`),
				lineRecord("worktree directory", `Worktree directory: .+`), lineRecord("prompt source", `Prompt source: .+`),
				lineRecord("issue fetch", `Fetching issue #[0-9]+ from .+`), lineRecord("title", `Title: .+`),
				lineRecord("labels", `Labels: .+`), branchRecord(), lineRecord("worktree path", `Path: .+`),
				lineRecord("base branch", `Base:.*`), lineRecord("worktree creation", `\[DRY-RUN\] Would run: git worktree add .+`),
				lineRecord("ready worktree", `Worktree ready at: .+`),
			},
			setup: func(t *testing.T, fixture *parityFixture) {
				config := filepath.Join(fixture.repo, ".start-issue")
				if err := os.MkdirAll(config, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(config, "agent"), []byte("none\n"), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(fixture.home, ".config", "start-issue", "agent"), []byte("codex\n"), 0644); err != nil {
					t.Fatal(err)
				}
			},
			strict: true,
		},
		{
			name: "dry-run", args: []string{"1", "--agent", "none", "--no-init", "--dry-run"}, strict: true,
			records: []parityOutputRecord{
				lineRecord("agent", `Agent: .+`), lineRecord("agent source", `Agent source: .+`),
				lineRecord("worktree directory", `Worktree directory: .+`), lineRecord("prompt source", `Prompt source: .+`),
				lineRecord("issue fetch", `Fetching issue #[0-9]+ from .+`), lineRecord("title", `Title: .+`),
				lineRecord("labels", `Labels: .+`), branchRecord(), lineRecord("worktree path", `Path: .+`),
				lineRecord("base branch", `Base:.*`), lineRecord("worktree creation", `\[DRY-RUN\] Would run: git worktree add .+`),
				lineRecord("ready worktree", `Worktree ready at: .+`),
			},
		},
		{
			name: "worktree-creation", args: []string{"1", "--agent", "none", "--no-init"}, strict: true,
			records: []parityOutputRecord{
				lineRecord("agent", `Agent: .+`), lineRecord("agent source", `Agent source: .+`),
				lineRecord("worktree directory", `Worktree directory: .+`), lineRecord("prompt source", `Prompt source: .+`),
				lineRecord("issue fetch", `Fetching issue #[0-9]+ from .+`), lineRecord("title", `Title: .+`),
				lineRecord("labels", `Labels: .+`), branchRecord(), lineRecord("worktree path", `Path: .+`),
				lineRecord("base branch", `Base:.*`), lineRecord("ready worktree", `Worktree ready at: .+`),
			},
		},
		{
			name:  "branch-worktree-reuse",
			args:  []string{"1", "--agent", "none", "--no-init"},
			input: "1\n",
			records: []parityOutputRecord{
				lineRecord("agent", `Agent: .+`), lineRecord("agent source", `Agent source: .+`),
				lineRecord("worktree directory", `Worktree directory: .+`), lineRecord("prompt source", `Prompt source: .+`),
				lineRecord("issue fetch", `Fetching issue #[0-9]+ from .+`), branchRecord(), lineRecord("worktree path", `Path: .+`),
				lineRecord("existing worktree", `Existing worktree: .+`), lineRecord("ready worktree", `Worktree ready at: .+`),
			},
			setup: func(t *testing.T, fixture *parityFixture) {
				fixture.branchExists = true
				if err := os.MkdirAll(fixture.fakeWorktree(), 0755); err != nil {
					t.Fatal(err)
				}
			},
			strict: true,
		},
		{
			name:  "branch-conflict-suffix",
			args:  []string{"1", "--agent", "none", "--no-init"},
			input: "2\n",
			records: []parityOutputRecord{
				lineRecord("agent", `Agent: .+`), lineRecord("worktree directory", `Worktree directory: .+`),
				lineRecord("issue fetch", `Fetching issue #[0-9]+ from .+`), branchRecord(), lineRecord("worktree path", `Path: .+`),
				lineRecord("new branch", `New branch name: .+`), lineRecord("ready worktree", `Worktree ready at: .+`),
			},
			setup: func(t *testing.T, fixture *parityFixture) {
				fixture.branchExists = true
			},
			strict: true,
		},
		{
			name:  "worktree-path-reuse",
			args:  []string{"1", "--agent", "none", "--no-init"},
			input: "1\n",
			records: []parityOutputRecord{
				lineRecord("agent", `Agent: .+`), lineRecord("worktree directory", `Worktree directory: .+`),
				lineRecord("issue fetch", `Fetching issue #[0-9]+ from .+`), branchRecord(), lineRecord("worktree path", `Path: .+`),
				lineRecord("registered branch", `Registered branch: .+`), lineRecord("ready worktree", `Worktree ready at: .+`),
			},
			setup: func(t *testing.T, fixture *parityFixture) {
				fixture.pathBranch = true
				if err := os.MkdirAll(fixture.fakeWorktree(), 0755); err != nil {
					t.Fatal(err)
				}
			},
			strict: true,
		},
		{
			name:  "worktree-path-conflict-dry-run",
			args:  []string{"1", "--agent", "none", "--no-init", "--dry-run"},
			input: "1\n",
			// ID-01 is the approved dry-run safety difference documented in
			// memory-bank/features/FT-016/decision-log.md.
			intentionalDifferenceID: "ID-01",
			setup: func(t *testing.T, fixture *parityFixture) {
				fixture.pathBranch = true
				if err := os.MkdirAll(fixture.fakeWorktree(), 0755); err != nil {
					t.Fatal(err)
				}
			},
			strict: false,
			// Bash consumed the supplied choice in dry-run. The Go command's
			// documented safety contract instead reports that it would prompt,
			// without selecting a destructive or reuse action.
			assertBaseline: func(t *testing.T, result parityResult) {
				assertParityOutputContains(t, result, "Worktree path already exists")
			},
			assertGo: func(t *testing.T, result parityResult) {
				assertParityOutputContains(t, result, "Worktree path exists; would prompt for reuse or delete/recreate")
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			bashFixture, goFixture := newParityFixture(t), newParityFixture(t)
			if test.setup != nil {
				test.setup(t, &bashFixture)
				test.setup(t, &goFixture)
			}
			if test.intentionalDifferenceID != "" && (test.strict || test.assertBaseline == nil || test.assertGo == nil) {
				t.Fatalf("intentional difference %s must declare distinct baseline and Go expectations", test.intentionalDifferenceID)
			}
			want := bashFixture.runBaseline(t, baseline, test.input, test.args...)
			got := goFixture.runResult(test.input, test.args...)
			if test.strict {
				assertParityResult(t, want, got, test.records)
			}
			if test.intentionalDifferenceID != "" {
				assertParitySideEffects(t, want, got)
			}
			if test.assertOutcome != nil {
				t.Run("bash-contract", func(t *testing.T) { test.assertOutcome(t, want) })
				t.Run("go-contract", func(t *testing.T) { test.assertOutcome(t, got) })
			}
			if test.assertBaseline != nil {
				t.Run("bash-contract", func(t *testing.T) { test.assertBaseline(t, want) })
			}
			if test.assertGo != nil {
				t.Run("go-contract", func(t *testing.T) { test.assertGo(t, got) })
			}
		})
	}

	for _, agent := range []string{"claude", "codex", "kimi", "pi", "none"} {
		t.Run("agent-launch-"+agent, func(t *testing.T) {
			bashFixture, goFixture := newParityFixture(t), newParityFixture(t)
			args := []string{"1", "--agent", agent, "--model", "fixture-model", "--no-init", "--dry-run"}
			want := bashFixture.runBaseline(t, baseline, "", args...)
			got := goFixture.runResult("", args...)
			assertAgentLaunchParity(t, want, got, agent)
			assertAgentLaunchContract(t, want, agent)
			assertAgentLaunchContract(t, got, agent)
		})
	}
}

type parityResult struct {
	exitCode   int
	rawOutput  string
	gitLog     string
	ghLog      string
	filesystem []string
}

// parityOutputRecord belongs to one named parity case. It compares a
// user-observable record that is stable across runtimes while letting the case
// retain its own output contract. We do not use a global output whitelist:
// every strict case explicitly declares the records it must preserve.
type parityOutputRecord struct {
	name      string
	pattern   *regexp.Regexp
	normalize func(string) string
}

func lineRecord(name, expression string) parityOutputRecord {
	return parityOutputRecord{name: name, pattern: regexp.MustCompile(expression)}
}

func branchRecord() parityOutputRecord {
	return parityOutputRecord{
		name:    "branch",
		pattern: regexp.MustCompile(`Branch: .+`),
		normalize: func(record string) string {
			// Only Bash's elapsed-time decoration differs; the branch and its
			// source remain part of the comparison.
			return regexp.MustCompile(` \([0-9]+s, `).ReplaceAllString(record, " (")
		},
	}
}

func shellEscapedLineRecord(name, expression string) parityOutputRecord {
	return parityOutputRecord{
		name:    name,
		pattern: regexp.MustCompile(expression),
		// Bash shell-escapes the placeholder in its displayed command, while
		// Go prints the rendered prompt directly. The placeholder contents are
		// the contract; its command-display quoting is not.
		normalize: func(record string) string {
			return strings.ReplaceAll(record, `\`, "")
		},
	}
}

func assertParityOutputContains(t *testing.T, result parityResult, values ...string) {
	t.Helper()
	for _, value := range values {
		if !strings.Contains(result.rawOutput, value) {
			t.Fatalf("output missing %q:\n%s", value, result.rawOutput)
		}
	}
}

func assertAgentLaunchContract(t *testing.T, result parityResult, agent string) {
	t.Helper()
	assertParityOutputContains(t, result, "Agent: "+agent)
	if agent == "none" {
		assertParityOutputContains(t, result, "Worktree ready at:", "Suggested agent commands:")
		return
	}
	commands := map[string]string{
		"claude": "claude --model fixture-model --dangerously-skip-permissions",
		"codex":  "codex --model fixture-model --cd",
		"kimi":   "kimi --model fixture-model --work-dir",
		"pi":     "pi --model fixture-model",
	}
	assertParityOutputContains(t, result, "[DRY-RUN] Would run:", commands[agent])
}

func assertAgentLaunchParity(t *testing.T, want, got parityResult, agent string) {
	t.Helper()
	if want.exitCode != got.exitCode {
		t.Fatalf("%s launch exit code differs: Bash=%d Go=%d", agent, want.exitCode, got.exitCode)
	}
	if bashRecord, goRecord := agentLaunchRecord(want.rawOutput, agent), agentLaunchRecord(got.rawOutput, agent); bashRecord != goRecord {
		t.Fatalf("%s launch contract differs:\nBash:\n%s\nGo:\n%s", agent, bashRecord, goRecord)
	}
}

func agentLaunchRecord(output, agent string) string {
	var records []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Agent: ") {
			records = append(records, line)
		}
		if agent != "none" && strings.Contains(line, "[DRY-RUN] Would run:") && strings.Contains(line, agent+" ") {
			records = append(records, launchAdapterPrefix(line, agent))
		}
		if agent == "none" && (strings.Contains(line, "Worktree ready at:") || line == "Suggested agent commands:") {
			records = append(records, line)
		}
	}
	return strings.Join(records, "\n")
}

func launchAdapterPrefix(command, agent string) string {
	markers := map[string]string{
		"claude": "--dangerously-skip-permissions",
		"codex":  "--dangerously-bypass-approvals-and-sandbox",
		"kimi":   "--yolo -p",
		"pi":     "pi --model fixture-model",
	}
	marker := markers[agent]
	if index := strings.Index(command, marker); index >= 0 {
		return command[:index+len(marker)]
	}
	return command
}

func extractBashParityBaseline(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the Bash baseline oracle is POSIX-only")
	}
	root := repoRoot(t)
	return filepath.Join(root, "cmd", "start-issue", "testdata", "bash-v1", "scripts", "start-issue")
}

func assertParityResult(t *testing.T, want, got parityResult, records []parityOutputRecord) {
	t.Helper()
	if want.exitCode != got.exitCode {
		t.Fatalf("exit code differs: Bash=%d Go=%d\nBash output:\n%s\nGo output:\n%s", want.exitCode, got.exitCode, want.rawOutput, got.rawOutput)
	}
	if len(records) == 0 {
		t.Fatal("strict parity case has no declared output records")
	}
	for _, record := range records {
		wantRecord := parityRecord(t, want.rawOutput, record)
		gotRecord := parityRecord(t, got.rawOutput, record)
		if wantRecord != gotRecord {
			t.Fatalf("%s output differs:\nBash: %s\nGo: %s", record.name, wantRecord, gotRecord)
		}
	}
	assertParitySideEffects(t, want, got)
}

func assertParitySideEffects(t *testing.T, want, got parityResult) {
	t.Helper()
	if want.gitLog != got.gitLog || want.ghLog != got.ghLog {
		t.Fatalf("fake command logs differ:\nBash git=%q gh=%q\nGo git=%q gh=%q", want.gitLog, want.ghLog, got.gitLog, got.ghLog)
	}
	if strings.Join(want.filesystem, "\n") != strings.Join(got.filesystem, "\n") {
		t.Fatalf("filesystem state differs:\nBash=%v\nGo=%v", want.filesystem, got.filesystem)
	}
}

func parityRecord(t *testing.T, output string, record parityOutputRecord) string {
	t.Helper()
	match := record.pattern.FindString(output)
	if match == "" {
		t.Fatalf("%s record missing from output:\n%s", record.name, output)
	}
	match = strings.TrimSpace(match)
	if record.normalize != nil {
		match = record.normalize(match)
	}
	return match
}

func TestCLIParityDryRunResolvesBranchConflicts(t *testing.T) {
	baseline := extractBashParityBaseline(t)
	for _, test := range []struct {
		name, input string
		setup       func(*testing.T, *parityFixture)
		records     []parityOutputRecord
	}{
		{
			name:  "reuse attached worktree",
			input: "1\n",
			setup: func(t *testing.T, fixture *parityFixture) {
				fixture.branchExists = true
				if err := os.MkdirAll(fixture.fakeWorktree(), 0755); err != nil {
					t.Fatal(err)
				}
			},
			records: []parityOutputRecord{
				lineRecord("existing worktree", `Existing worktree: .+`),
				lineRecord("ready worktree", `Worktree ready at: .+`),
			},
		},
		{
			name:  "suffix branch",
			input: "2\n",
			setup: func(_ *testing.T, fixture *parityFixture) {
				fixture.branchExists = true
			},
			records: []parityOutputRecord{
				lineRecord("new branch", `New branch name: .+`),
				lineRecord("worktree creation", `\[DRY-RUN\] Would run: git worktree add .+`),
				lineRecord("ready worktree", `Worktree ready at: .+`),
			},
		},
		{
			name:  "delete and recreate",
			input: "3\n",
			setup: func(_ *testing.T, fixture *parityFixture) {
				fixture.branchExists = true
			},
			records: []parityOutputRecord{
				lineRecord("worktree creation", `\[DRY-RUN\] Would run: git worktree add .+`),
				lineRecord("ready worktree", `Worktree ready at: .+`),
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			bashFixture, goFixture := newParityFixture(t), newParityFixture(t)
			test.setup(t, &bashFixture)
			test.setup(t, &goFixture)
			args := []string{"1", "--agent", "none", "--no-init", "--dry-run"}
			want := bashFixture.runBaseline(t, baseline, test.input, args...)
			got := goFixture.runResult(test.input, args...)
			assertParityResult(t, want, got, test.records)
		})
	}
}

func TestCLIAIBranchReportsOneAIRecord(t *testing.T) {
	fixture := newParityFixture(t)
	writeExecutable(t, filepath.Join(fixture.bin, "pi"), "#!/bin/sh\nprintf '%s\\n' feature/issue-1-add-login-button\n")

	output, err := fixture.run("", "1", "--agent", "pi", "--ai", "--no-init", "--dry-run")
	if err != nil {
		t.Fatalf("AI dry-run failed: %v\n%s", err, output)
	}
	branchRecords := 0
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == "Branch: feature/issue-1-add-login-button (ai:pi)" {
			branchRecords++
		}
	}
	if branchRecords != 1 {
		t.Fatalf("AI branch was reported %d times:\n%s", branchRecords, output)
	}
	if !strings.Contains(output, "Branch: feature/issue-1-add-login-button (ai:pi)") || strings.Contains(output, "Branch: feature/issue-1-add-login-button (fast)") {
		t.Fatalf("AI branch source is contradictory:\n%s", output)
	}
}

func TestCLIParityInstallerWorkflow(t *testing.T) {
	root := repoRoot(t)
	dir := t.TempDir()
	asset := filepath.Join(dir, "start-issue")
	writeExecutable(t, asset, "#!/bin/sh\n[ \"$1\" = --version ] && echo 'start-issue v2.1.0'\n")
	checksum := sha256.Sum256([]byte("#!/bin/sh\n[ \"$1\" = --version ] && echo 'start-issue v2.1.0'\n"))
	manifest := filepath.Join(dir, "checksums.txt")
	if err := os.WriteFile(manifest, []byte(fmt.Sprintf("%x  start-issue\n", checksum)), 0644); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "bin", "start-issue")
	command := exec.Command(requireBash(t), filepath.Join(root, "install.sh"))
	command.Env = append(os.Environ(),
		"START_ISSUE_ASSET_URL=file://"+asset,
		"START_ISSUE_CHECKSUM_URL=file://"+manifest,
		"TARGET="+target,
		"BINDIR="+filepath.Dir(target),
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("installer failed: %v\n%s", err, output)
	}
	versionOutput, err := exec.Command(target, "--version").CombinedOutput()
	if err != nil || strings.TrimSpace(string(versionOutput)) != "start-issue v2.1.0" {
		t.Fatalf("installed version = %q, %v", versionOutput, err)
	}
}

func TestInstallerChecksumNameExcludesURLQuery(t *testing.T) {
	bash := requireBash(t)
	assetName, err := releaseAssetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Skipf("unsupported test platform: %v", err)
	}
	asset := []byte("#!/bin/sh\n[ \"$1\" = --version ] && echo 'start-issue v2.1.0'\n")
	checksum := fmt.Sprintf("%x  %s\n", sha256.Sum256(asset), assetName)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("local HTTP listener unavailable: %v", err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/" + assetName:
			if request.URL.Query().Get("token") != "presigned" {
				http.Error(response, "missing token", http.StatusBadRequest)
				return
			}
			_, _ = response.Write(asset)
		case "/checksums.txt":
			_, _ = response.Write([]byte(checksum))
		default:
			http.NotFound(response, request)
		}
	}))
	server.Listener = listener
	server.Start()
	defer server.Close()

	target := filepath.Join(t.TempDir(), "bin", "start-issue")
	command := exec.Command(bash, filepath.Join(repoRoot(t), "install.sh"))
	command.Env = append(os.Environ(),
		"START_ISSUE_ASSET_URL="+server.URL+"/"+assetName+"?token=presigned",
		"START_ISSUE_CHECKSUM_URL="+server.URL+"/checksums.txt",
		"TARGET="+target,
		"BINDIR="+filepath.Dir(target),
	)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("installer rejected the checksummed presigned asset URL: %v\n%s", err, output)
	}
}

func TestInstallerDefaultsEachURLIndependently(t *testing.T) {
	bash := requireBash(t)
	root := repoRoot(t)
	bin := t.TempDir()
	writeExecutable(t, filepath.Join(bin, "curl"), "#!/bin/sh\nexit 1\n")
	assetName, err := releaseAssetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Skipf("unsupported test platform: %v", err)
	}
	defaultAsset := "https://github.com/dapi/start-issue/releases/latest/download/" + assetName
	defaultChecksum := "https://github.com/dapi/start-issue/releases/latest/download/checksums.txt"

	for _, test := range []struct {
		name, asset, checksum, wantAsset, wantChecksum string
	}{
		{
			name:         "custom asset keeps default checksum",
			asset:        "https://example.test/custom/start-issue",
			wantAsset:    "https://example.test/custom/start-issue",
			wantChecksum: defaultChecksum,
		},
		{
			name:         "custom checksum keeps default asset",
			checksum:     "https://example.test/custom/checksums.txt",
			wantAsset:    defaultAsset,
			wantChecksum: "https://example.test/custom/checksums.txt",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			command := exec.Command(bash, filepath.Join(root, "install.sh"), "--debug")
			command.Env = append(os.Environ(),
				"START_ISSUE_ASSET_URL="+test.asset,
				"START_ISSUE_CHECKSUM_URL="+test.checksum,
				"BINDIR="+t.TempDir(),
				"TARGET="+filepath.Join(t.TempDir(), "start-issue"),
				"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
			)
			output, err := command.CombinedOutput()
			if err == nil {
				t.Fatal("installer unexpectedly succeeded with a failing curl fixture")
			}
			for _, want := range []string{"Asset URL: " + test.wantAsset, "Checksum URL: " + test.wantChecksum} {
				if !strings.Contains(string(output), want) {
					t.Fatalf("installer did not resolve %q:\n%s", want, output)
				}
			}
		})
	}
}

func requireBash(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("install.sh is a POSIX Bash installer")
	}
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("Bash is unavailable; installer integration test is not applicable")
	}
	return bash
}

func TestCLIParityUpdateWorkflowPreservesInvocationSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows update is intentionally manual")
	}
	root := repoRoot(t)
	dir, bin := t.TempDir(), t.TempDir()
	target := filepath.Join(dir, "start-issue-target")
	invocation := filepath.Join(dir, "start-issue")
	build := exec.Command("go", "build", "-o", target, "./cmd/start-issue")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build update fixture: %v\n%s", err, output)
	}
	if err := os.Symlink(target, invocation); err != nil {
		t.Fatal(err)
	}
	assetName, err := releaseAssetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Skipf("unsupported test platform: %v", err)
	}
	asset := []byte("#!/bin/sh\nif [ \"$1\" = --version ]; then echo 'start-issue v2.1.0'; exit 0; fi\n")
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
	release := filepath.Join(dir, "release.json")
	metadata := fmt.Sprintf(`{"tag_name":"v2.1.0","assets":[{"name":"%s","browser_download_url":"%s/asset"},{"name":"checksums.txt","browser_download_url":"%s/checksums"}]}`, assetName, server.URL, server.URL)
	if err := os.WriteFile(release, []byte(metadata), 0644); err != nil {
		t.Fatal(err)
	}
	writeExecutable(t, filepath.Join(bin, "gh"), "#!/bin/sh\nif [ \"$1\" = auth ]; then exit 0; fi\ncat \"$START_ISSUE_FAKE_RELEASE\"\n")
	command := exec.Command(invocation, "update")
	command.Env = append(os.Environ(), "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"), "START_ISSUE_FAKE_RELEASE="+release)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("update failed: %v\n%s", err, output)
	}
	info, err := os.Lstat(invocation)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("update replaced invocation symlink %s instead of its target", invocation)
	}
	versionOutput, err := exec.Command(invocation, "--version").CombinedOutput()
	if err != nil || strings.TrimSpace(string(versionOutput)) != "start-issue v2.1.0" {
		t.Fatalf("updated version = %q, %v", versionOutput, err)
	}
}

func TestV1UpdateBridgeInstallsVerifiedPlatformBinary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the v1 update bridge is a POSIX compatibility asset")
	}
	root, dir := repoRoot(t), t.TempDir()
	assetName, err := releaseAssetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Skipf("unsupported test platform: %v", err)
	}
	asset := filepath.Join(dir, assetName)
	assetBody := []byte("#!/bin/sh\n[ \"$1\" = --version ] && echo 'start-issue v2.0.0'\n")
	if err := os.WriteFile(asset, assetBody, 0755); err != nil {
		t.Fatal(err)
	}
	checksums := filepath.Join(dir, "checksums.txt")
	if err := os.WriteFile(checksums, []byte(fmt.Sprintf("%x  %s\n", sha256.Sum256(assetBody), assetName)), 0644); err != nil {
		t.Fatal(err)
	}
	bridge := filepath.Join(dir, "start-issue")
	body, err := os.ReadFile(filepath.Join(root, "scripts", "v1-upgrade-shim"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bridge, body, 0755); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(requireBash(t), bridge, "--version")
	command.Env = append(os.Environ(), "START_ISSUE_UPGRADE_ASSET_URL=file://"+asset, "START_ISSUE_UPGRADE_CHECKSUM_URL=file://"+checksums)
	output, err := command.CombinedOutput()
	if err != nil || strings.TrimSpace(string(output)) != "start-issue v2.0.0" {
		t.Fatalf("v1 bridge = %q, %v", output, err)
	}
	installed, err := os.ReadFile(bridge)
	if err != nil || string(installed) != string(assetBody) {
		t.Fatalf("v1 bridge did not replace itself with the verified binary: %v", err)
	}
}

func TestGoReleaserBridgeChecksumNamesPublishedAsset(t *testing.T) {
	config, err := os.ReadFile(filepath.Join(repoRoot(t), ".goreleaser.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(config), "(cd .release && sha256sum start-issue > start-issue.sha256)") {
		t.Fatal("v1 bridge checksum must name start-issue, the published asset")
	}
}

type parityFixture struct {
	home, repo, bin, worktrees, gitLog, ghLog, initMarker string
	issueLabel                                            string
	branchExists, pathBranch                              bool
}

func (fixture parityFixture) fakeWorktree() string {
	return canonicalPath(filepath.Join(fixture.worktrees, "feature", "issue-1-add-login-button"))
}

func newParityFixture(t *testing.T) parityFixture {
	t.Helper()
	fixture := parityFixture{home: t.TempDir(), repo: t.TempDir(), bin: t.TempDir(), worktrees: t.TempDir(), gitLog: filepath.Join(t.TempDir(), "git.log"), ghLog: filepath.Join(t.TempDir(), "gh.log")}
	if err := os.MkdirAll(filepath.Join(fixture.home, ".config", "start-issue"), 0755); err != nil {
		t.Fatal(err)
	}
	writeExecutable(t, filepath.Join(fixture.bin, "git"), `#!/bin/sh
if [ "$1" = "-C" ]; then
  printf '%s\n' "$START_ISSUE_TEST_REPO"
  exit 0
fi
case "$1 $2 $3" in
  "rev-parse --show-toplevel ") printf '%s\n' "$START_ISSUE_TEST_REPO" ;;
  "remote get-url origin") printf '%s\n' 'git@github.com:owner/repo.git' ;;
  "symbolic-ref refs/remotes/origin/HEAD") printf '%s\n' 'refs/remotes/origin/master' ;;
  show-ref\ --verify\ *) [ "${START_ISSUE_FAKE_BRANCH_EXISTS:-}" = 1 ] && [ "$4" = "refs/heads/feature/issue-1-add-login-button" ] && exit 0; exit 1 ;;
  "worktree list --porcelain")
    if [ "${START_ISSUE_FAKE_BRANCH_EXISTS:-}" = 1 ] || [ "${START_ISSUE_FAKE_PATH_BRANCH:-}" = 1 ]; then
      printf 'worktree %s\nbranch refs/heads/feature/issue-1-add-login-button\n' "$START_ISSUE_FAKE_WORKTREE"
    fi
    ;;
  "fetch origin master") ;;
  "worktree add -b")
    mkdir -p "$5"
    if [ -n "${START_ISSUE_INIT_MARKER:-}" ]; then
      printf '%s\n' '#!/bin/sh' 'touch "$START_ISSUE_INIT_MARKER"' > "$5/init.sh"
      chmod +x "$5/init.sh"
    fi
    printf '%s\n' "$*" >> "$START_ISSUE_GIT_LOG"
    ;;
esac
`)
	writeExecutable(t, filepath.Join(fixture.bin, "gh"), `#!/bin/sh
printf '%s\n' "$*" >> "$START_ISSUE_GH_LOG"
if [ "$1" = auth ] && [ "$2" = status ]; then exit 0; fi
if [ "$1" = api ]; then
  printf '{"title":"Add login button","body":"Fixture body","labels":[{"name":"%s"}]}\n' "${START_ISSUE_FAKE_LABEL:-feature}"
fi
`)
	writeExecutable(t, filepath.Join(fixture.bin, "jq"), `#!/bin/sh
case "$*" in
  *".title"*) printf '%s\n' 'Add login button' ;;
  *".body"*) printf '%s\n' 'Fixture body' ;;
  *".labels"*) printf '%s\n' "${START_ISSUE_FAKE_LABEL:-feature}" ;;
esac
`)
	writeExecutable(t, filepath.Join(fixture.bin, "zellij-tab-status"), "#!/bin/sh\nexit 0\n")
	return fixture
}

func (fixture parityFixture) run(input string, args ...string) (string, error) {
	command := exec.Command(os.Args[0], append([]string{"-test.run=^TestCLIParityHelper$", "--"}, args...)...)
	return fixture.runCommand(command, input)
}

func (fixture parityFixture) runResult(input string, args ...string) parityResult {
	output, err := fixture.run(input, args...)
	return fixture.result(output, err)
}

func (fixture parityFixture) runBaseline(t *testing.T, baseline, input string, args ...string) parityResult {
	command := exec.Command(requireBash(t), append([]string{baseline}, args...)...)
	output, err := fixture.runCommand(command, input)
	return fixture.result(output, err)
}

func (fixture parityFixture) result(output string, err error) parityResult {
	gitLog, _ := os.ReadFile(fixture.gitLog)
	ghLog, _ := os.ReadFile(fixture.ghLog)
	return parityResult{
		exitCode:   commandExitCode(err),
		rawOutput:  normalizeRawParityOutput(output, fixture),
		gitLog:     normalizeRawParityOutput(string(gitLog), fixture),
		ghLog:      normalizeRawParityOutput(string(ghLog), fixture),
		filesystem: parityFilesystem(fixture),
	}
}

func normalizeRawParityOutput(output string, fixture parityFixture) string {
	stripANSI := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	output = stripANSI.ReplaceAllString(output, "")
	// Preserve which semantic root owns a path. In particular, the worktree
	// parent can be located inside the repository or home directory, so its
	// more-specific path must be substituted before its containing root.
	for _, root := range parityRoots(fixture) {
		for _, path := range root.paths() {
			output = strings.ReplaceAll(output, path, root.token)
		}
	}
	return output
}

func commandExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exited *exec.ExitError
	if errors.As(err, &exited) {
		return exited.ExitCode()
	}
	return -1
}

func parityFilesystem(fixture parityFixture) []string {
	var paths []string
	for _, root := range parityRoots(fixture) {
		_ = filepath.Walk(root.path, func(path string, info os.FileInfo, err error) error {
			if err != nil || path == root.path {
				return err
			}
			rel, relErr := filepath.Rel(root.path, path)
			if relErr == nil {
				if info.IsDir() {
					rel += "/"
				}
				paths = append(paths, root.token+"/"+filepath.ToSlash(rel))
			}
			return nil
		})
	}
	sort.Strings(paths)
	return paths
}

type parityRoot struct {
	token, path string
}

func (root parityRoot) paths() []string {
	paths := []string{canonicalPath(root.path), root.path}
	sort.SliceStable(paths, func(i, j int) bool { return len(paths[i]) > len(paths[j]) })
	return uniqueStrings(paths)
}

func parityRoots(fixture parityFixture) []parityRoot {
	roots := []parityRoot{
		{token: "<HOME>", path: fixture.home},
		{token: "<REPO>", path: fixture.repo},
		{token: "<WORKTREES>", path: fixture.worktrees},
	}
	// Nested roots require the longest source paths to be normalized first;
	// otherwise replacing a parent would erase the child's identity.
	sort.SliceStable(roots, func(i, j int) bool {
		return len(canonicalPath(roots[i].path)) > len(canonicalPath(roots[j].path))
	})
	return roots
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func (fixture parityFixture) runCommand(command *exec.Cmd, input string) (string, error) {
	command.Dir = fixture.repo
	command.Stdin = strings.NewReader(input)
	command.Env = append(os.Environ(),
		"START_ISSUE_PARITY_HELPER=1",
		"HOME="+fixture.home,
		"PATH="+fixture.bin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"START_ISSUE_TEST_REPO="+fixture.repo,
		"START_ISSUE_GIT_LOG="+fixture.gitLog,
		"START_ISSUE_GH_LOG="+fixture.ghLog,
		"START_ISSUE_WORKTREE_DIR="+fixture.worktrees,
		"START_ISSUE_INIT_MARKER="+fixture.initMarker,
	)
	if fixture.branchExists {
		command.Env = append(command.Env, "START_ISSUE_FAKE_BRANCH_EXISTS=1")
	}
	if fixture.pathBranch {
		command.Env = append(command.Env, "START_ISSUE_FAKE_PATH_BRANCH=1")
	}
	if fixture.branchExists || fixture.pathBranch {
		command.Env = append(command.Env, "START_ISSUE_FAKE_WORKTREE="+fixture.fakeWorktree())
	}
	if fixture.issueLabel != "" {
		command.Env = append(command.Env, "START_ISSUE_FAKE_LABEL="+fixture.issueLabel)
	}
	output, err := command.CombinedOutput()
	return string(output), err
}

func repoRoot(t *testing.T) string {
	t.Helper()
	current, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(filepath.Join(current, "..", ".."))
}
