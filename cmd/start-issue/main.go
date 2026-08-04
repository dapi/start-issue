package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"
)

// version is set by release builds with -ldflags. go install records the
// module version in build metadata instead, which runningVersion reads below.
var version string

// sourceVersion identifies direct development builds that were not made by
// Make or go install. Release and source builds inject/record their version.
const sourceVersion = "dev"

const defaultReleaseRepository = "dapi/start-issue"

var gitDescribeSuffix = regexp.MustCompile(`^(.*?)(?:-[0-9]+-g[0-9a-fA-F]+(?:-dirty)?|-dirty)$`)

func runningVersion() string {
	if version != "" {
		return strings.TrimPrefix(version, "v")
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return sourceVersion
	}
	return versionFromBuildInfo(info)
}

func versionFromBuildInfo(info *debug.BuildInfo, fallback ...string) string {
	defaultVersion := sourceVersion
	if len(fallback) > 0 {
		defaultVersion = fallback[0]
	}
	if info != nil && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return strings.TrimPrefix(info.Main.Version, "v")
	}
	return defaultVersion
}

type options struct {
	repo, base, worktreeDir, agent, model, promptFile, prompt, command       string
	promptOutput, worktreeDirSource                                          string
	issue                                                                    string
	dryRun, noInit, flat, ai, improvePrompt, humanGate, project, user, force bool
	mode                                                                     string
}

type issue struct {
	Title, Body string
	Labels      []issueLabel
}

type issueLabel struct {
	Name string `json:"name"`
}

// UnmarshalJSON preserves the shell implementation's treatment of a GitHub
// issue with no body: GitHub represents it as null, while the CLI uses an
// empty string when rendering the prompt.
func (in *issue) UnmarshalJSON(data []byte) error {
	var value struct {
		Title  string       `json:"title"`
		Body   *string      `json:"body"`
		Labels []issueLabel `json:"labels"`
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	in.Title = value.Title
	in.Body = ""
	if value.Body != nil {
		in.Body = *value.Body
	}
	in.Labels = value.Labels
	return nil
}

type exitError struct {
	code int
	err  error
}

func (e exitError) Error() string { return e.err.Error() }

func (e exitError) Unwrap() error { return e.err }

type promptFileNotFoundError struct {
	path  string
	cause error
}

func (e promptFileNotFoundError) Error() string {
	return fmt.Sprintf("Prompt file not found: %s", e.path)
}

func (e promptFileNotFoundError) Unwrap() error { return e.cause }

func main() {
	o, err := parse(os.Args[1:])
	if err != nil {
		die(err)
	}
	printBanner()
	if o.mode != "" {
		if err := runMode(o); err != nil {
			die(err)
		}
		return
	}
	if o.issue == "" {
		if err := runMissingIssue(o); err != nil {
			die(err)
		}
		return
	}
	if err := run(o); err != nil {
		die(err)
	}
}

func parse(args []string) (options, error) {
	o := options{worktreeDir: os.Getenv("START_ISSUE_WORKTREE_DIR")}
	if o.worktreeDir != "" {
		o.worktreeDirSource = "START_ISSUE_WORKTREE_DIR"
	}
	var err error
	for len(args) > 0 {
		a := args[0]
		args = args[1:]
		value := func() (string, error) {
			if len(args) == 0 || strings.HasPrefix(args[0], "-") {
				return "", fmt.Errorf("%s requires a value.", a)
			}
			v := args[0]
			args = args[1:]
			if v == "" {
				return "", fmt.Errorf("%s requires a value.", a)
			}
			return v, nil
		}
		switch a {
		case "--help", "-h":
			usage()
			os.Exit(0)
		case "--version", "-v":
			fmt.Printf("start-issue v%s\n", runningVersion())
			os.Exit(0)
		case "--repo", "-r":
			o.repo, err = value()
		case "--base", "-b":
			o.base, err = value()
		case "--worktree-dir", "-w":
			o.worktreeDir, err = value()
			if err == nil {
				o.worktreeDirSource = "CLI"
			}
		case "--agent":
			o.agent, err = value()
		case "--model":
			o.model, err = value()
		case "--prompt-file":
			o.promptFile, err = value()
		case "--prompt":
			o.prompt, err = value()
		case "--command", "-c":
			o.command, err = value()
		case "--no-agent", "--no-claude":
			o.agent = "none"
		case "--dry-run":
			o.dryRun = true
		case "--no-init":
			o.noInit = true
		case "--flat":
			o.flat = true
		case "--ai":
			o.ai = true
		case "--improve-prompt":
			o.improvePrompt = true
		case "--prompt-output-file":
			o.promptOutput, err = value()
		case "--human-gate":
			o.humanGate = true
		case "--project":
			o.project = true
		case "--user":
			o.user = true
		case "--force":
			o.force = true
		case "init", "setup", "--setup", "update", "--update", "install", "--install":
			mode := strings.TrimLeft(a, "-")
			if o.mode != "" {
				return o, fmt.Errorf("Use only one command mode; got %s and %s.", o.mode, mode)
			}
			o.mode = mode
		case "--human-gate-help":
			humanGateHelp()
			os.Exit(0)
		default:
			if strings.HasPrefix(a, "-") {
				return o, fmt.Errorf("Unknown option: %s. Use --help for usage.", a)
			}
			if o.issue != "" {
				return o, fmt.Errorf("Unexpected argument: %s", a)
			}
			o.issue = a
		}
		if err != nil {
			return o, err
		}
	}
	if o.prompt != "" && o.promptFile != "" {
		return o, errors.New("Use either --prompt-file or --prompt, not both.")
	}
	if o.project && o.user {
		return o, errors.New("Use either --project or --user, not both.")
	}
	if (o.project || o.user || o.force) && o.mode != "init" {
		return o, errors.New("--project, --user, and --force are only valid with init.")
	}
	if o.mode != "" && o.issue != "" {
		return o, fmt.Errorf("Use either %s or <issue-url-or-number>, not both.", o.mode)
	}
	if o.worktreeDir == "" && o.mode == "" {
		home, err := userHomeDir()
		if err != nil {
			return o, fmt.Errorf("default worktree directory requires a home directory; set --worktree-dir or START_ISSUE_WORKTREE_DIR: %w", err)
		}
		o.worktreeDir = filepath.Join(home, "worktrees")
		o.worktreeDirSource = "built-in default"
	}
	return o, nil
}

func userHomeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(home) == "" || !filepath.IsAbs(home) {
		return "", errors.New("home directory is unavailable")
	}
	return home, nil
}

func run(o options) error {
	return runWithReader(o, bufio.NewReader(os.Stdin))
}

func runWithReader(o options, reader *bufio.Reader) error {
	onboardingErr := maybeRunFirstRunOnboarding(o.dryRun, o.command, reader)
	if err := need("git"); err != nil {
		return err
	}
	if onboardingErr != nil {
		return onboardingErr
	}
	root, err := output("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return errors.New("Not in a git repository")
	}
	root = strings.TrimSpace(root)
	agent, agentSource, err := resolveAgent(root, o.agent)
	if err != nil {
		return err
	}
	if o.humanGate && agent != "codex" {
		return fmt.Errorf("--human-gate requires agent 'codex'. Current agent: %s.", agent)
	}
	if o.improvePrompt && agent == "none" {
		return errors.New("--improve-prompt requires an agent. Use --agent claude, codex, kimi, or pi.")
	}
	if agent != "none" && !o.dryRun {
		if err := need(agent); err != nil {
			return fmt.Errorf("%s CLI not found. Install it or use --agent none.", agent)
		}
	}
	model, modelSource, err := resolveModel(root, o.model)
	if err != nil {
		return err
	}
	prompt, promptSource, promptLocation, promptFile, err := resolvePrompt(root, agent, o)
	if err != nil {
		return err
	}
	number, repo, err := parseIssue(o.issue, o.repo)
	if err != nil {
		return err
	}
	if repo == "" {
		repo, err = detectRepo()
		if err != nil {
			return err
		}
	}
	if o.base == "" {
		o.base = detectBase()
	}
	if err := checkGitHubAccess(); err != nil {
		return err
	}
	data, err := output("gh", "api", fmt.Sprintf("repos/%s/issues/%s", repo, number))
	if err != nil {
		return fmt.Errorf("Issue #%s not found in %s", number, repo)
	}
	var in issue
	if err = json.Unmarshal([]byte(data), &in); err != nil {
		return err
	}
	labels := []string{}
	for _, l := range in.Labels {
		labels = append(labels, l.Name)
	}
	if o.improvePrompt {
		return improvePrompt(root, agent, model, prompt, promptSource, promptFile, o, in, repo, number, strings.Join(labels, ", "))
	}
	renameZellijTab(number, o.dryRun)
	branch := branchName(number, in.Title, strings.Join(labels, ", "))
	branchSource := "fast"
	if o.ai {
		if generated, err := aiBranchName(agent, model, root, number, in.Title, strings.Join(labels, ", ")); err == nil && regexp.MustCompile(`^(feature|fix|hotfix|refactor|docs|test|chore)/issue-[0-9]+-[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$`).MatchString(generated) {
			branch = generated
			branchSource = "ai:" + agent
		} else {
			fmt.Printf("   Could not generate branch name with %s; using fast fallback\n", agent)
		}
	}
	issueURL := fmt.Sprintf("https://github.com/%s/issues/%s", repo, number)
	worktree := canonicalPath(worktreePath(o.worktreeDir, branch, o.flat))
	fmt.Printf("Agent: %s\nAgent source: %s\nModel: %s\nModel source: %s\nWorktree directory: %s (%s)\nPrompt source: %s\nPrompt location: %s\n\n", agent, agentSource, show(model), modelSource, o.worktreeDir, o.worktreeDirSource, promptSource, promptLocation)
	fmt.Printf("🔍 Fetching issue #%s from %s...\n   Title: %s\n", number, repo, in.Title)
	if len(labels) > 0 {
		fmt.Printf("   Labels: %s\n", strings.Join(labels, ", "))
	}
	fmt.Printf("   Branch: %s (%s)\n📁 Creating worktree...\n   Path: %s\n   Base: %s\n", branch, branchSource, worktree, o.base)
	branchExists := branchRefExists(branch)
	existing := branchWorktree(branch)
	recreate := false
	if branchExists {
		choice := branchConflictChoice(branch, existing, reader)
		switch choice {
		case "1":
			if existing == "" {
				return fmt.Errorf("No existing worktree found for branch '%s'. Use 3 to delete and recreate.", branch)
			}
			if err := validateReusedWorktree(existing, branch); err != nil {
				return err
			}
			if !o.noInit {
				runInit(existing, o.dryRun)
			}
			return launchSelected(o, agent, model, existing, renderIssuePrompt(prompt, issueURL, number, in, labels, repo, branch, existing, o.base))
		case "2":
			branch = nextSuffixedBranch(branch)
			worktree = canonicalPath(worktreePath(o.worktreeDir, branch, o.flat))
			fmt.Printf("   New branch name: %s\n", branch)
		case "3":
			if o.dryRun {
				if existing != "" {
					fmt.Printf("   [DRY-RUN] Would remove worktree: %s\n", existing)
				}
				fmt.Printf("   [DRY-RUN] Would delete branch: %s\n", branch)
				recreate = true
				break
			}
			if err := removeWorktreeAndBranch(existing, branch); err != nil {
				return err
			}
		default:
			return errors.New("Cancelled.")
		}
	}
	if !recreate {
		if _, err := os.Stat(worktree); err == nil {
			pathBranch, registered := worktreeRegistration(worktree)
			if o.dryRun {
				if !registered {
					fmt.Printf("   [DRY-RUN] Worktree path exists but is not a registered worktree; would stop and require manual recovery: %s\n", worktree)
					return nil
				}
				fmt.Printf("   [DRY-RUN] Worktree path exists; would prompt for reuse or delete/recreate: %s\n", worktree)
				return nil
			}
			choice, err := pathConflictChoice(worktree, pathBranch, registered, reader)
			if err != nil {
				return err
			}
			switch choice {
			case "1":
				if err := validateReusedWorktree(worktree, branch); err != nil {
					return err
				}
				if !o.noInit {
					runInit(worktree, o.dryRun)
				}
				return launchSelected(o, agent, model, worktree, renderIssuePrompt(prompt, issueURL, number, in, labels, repo, branch, worktree, o.base))
			case "2":
				if err := removeWorktreeAndBranch(worktree, strings.TrimPrefix(pathBranch, "refs/heads/")); err != nil {
					return err
				}
			default:
				return errors.New("Cancelled.")
			}
		}
	}
	rendered := renderIssuePrompt(prompt, issueURL, number, in, labels, repo, branch, worktree, o.base)
	if o.dryRun {
		fmt.Printf("   [DRY-RUN] Would run: git worktree add -b %s %s %s\n", branch, worktree, o.base)
		if o.humanGate {
			return humanGate(model, worktree, rendered, true)
		}
		return launchSelected(options{dryRun: true}, agent, model, worktree, rendered)
	}
	if err := os.MkdirAll(filepath.Dir(worktree), 0755); err != nil {
		return err
	}
	_ = command("git", "fetch", "origin", o.base, "--quiet")
	if err := command("git", "worktree", "add", "-b", branch, worktree, "origin/"+o.base); err != nil {
		if err = command("git", "worktree", "add", "-b", branch, worktree, o.base); err != nil {
			return errors.New("Failed to create worktree")
		}
	}
	if !o.noInit {
		runInit(worktree, false)
	}
	return launchSelected(o, agent, model, worktree, rendered)
}

func runInit(worktree string, dryRun bool) {
	if init := filepath.Join(worktree, "init.sh"); fileExists(init) {
		if dryRun {
			fmt.Printf("   [DRY-RUN] Would run: %s\n", init)
			return
		}
		if err := commandAt(worktree, "bash", "./init.sh"); err != nil {
			fmt.Println("Warning: init.sh exited with non-zero code")
		}
	}
}

func renameZellijTab(number string, dryRun bool) {
	if _, err := exec.LookPath("zellij-tab-status"); err != nil {
		if dryRun {
			fmt.Println("   [DRY-RUN] Would skip zellij tab rename: zellij-tab-status not found")
		}
		return
	}
	if dryRun {
		fmt.Printf("   [DRY-RUN] Would run: zellij-tab-status --set-name #%s\n", number)
		return
	}
	if err := command("zellij-tab-status", "--set-name", "#"+number); err != nil {
		fmt.Println("Warning: Could not rename zellij tab with zellij-tab-status")
	}
}

func fileExists(path string) bool { _, err := os.Stat(path); return err == nil }
func maybeRunFirstRunOnboarding(dryRun bool, command string, reader *bufio.Reader) error {
	home, err := userHomeDir()
	if err != nil {
		return nil
	}
	dir := filepath.Join(home, ".config", "start-issue")
	if _, err := os.Stat(dir); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	fmt.Println("Configuration is not initialized yet.")
	fmt.Println()
	fmt.Println("Usage: start-issue <issue-url-or-number> [options]")
	fmt.Println()
	fmt.Print("No start-issue user configuration found. Run setup now? [Y/n] ")
	if dryRun {
		fmt.Printf("[DRY-RUN] Would create first-run configuration marker: %s\n", dir)
		return nil
	}
	answer, err := reader.ReadString('\n')
	if err != nil {
		return errors.New("No response received.")
	}
	switch strings.TrimSpace(strings.ToLower(answer)) {
	case "", "y", "yes":
		return setupMode(home, false, command, reader)
	case "n", "no":
		return os.MkdirAll(dir, 0755)
	default:
		return fmt.Errorf("Invalid response: %s. Use y or n.", strings.TrimSpace(answer))
	}
}

func runMissingIssue(o options) error {
	if o.improvePrompt {
		return errors.New("--improve-prompt requires <issue-url-or-number>. Example: start-issue 123 --improve-prompt")
	}
	root, _ := output("git", "rev-parse", "--show-toplevel")
	root = strings.TrimSpace(root)
	if err := maybeRunFirstRunOnboarding(o.dryRun, o.command, bufio.NewReader(os.Stdin)); err != nil {
		return err
	}
	agent, agentSource, err := resolveAgent(root, o.agent)
	if err != nil {
		return err
	}
	model, modelSource, err := resolveModel(root, o.model)
	if err != nil {
		return err
	}
	_, promptSource, promptLocation, _, err := resolvePrompt(root, agent, o)
	if err != nil {
		return err
	}
	usage()
	fmt.Println("Current configuration:")
	fmt.Printf("  Agent: %s\n  Agent source: %s\n  Model: %s\n  Model source: %s\n  Prompt source: %s\n  Prompt location: %s\n  Worktree dir: %s (%s)\n", agent, agentSource, show(model), modelSource, promptSource, promptLocation, o.worktreeDir, o.worktreeDirSource)
	return errors.New("missing issue URL or issue number")
}

func runMode(o options) error {
	if o.mode == "update" {
		return updateMode(o)
	}
	root, _ := output("git", "rev-parse", "--show-toplevel")
	root = strings.TrimSpace(root)
	if o.mode == "install" {
		return installMode(o.dryRun)
	}
	if o.mode == "setup" {
		home, err := userHomeDir()
		if err != nil {
			return err
		}
		return setupMode(home, o.dryRun, o.command, bufio.NewReader(os.Stdin))
	}
	if o.mode != "init" {
		return fmt.Errorf("Unknown command mode: %s", o.mode)
	}
	home := ""
	if !o.project {
		var err error
		home, err = userHomeDir()
		if err != nil {
			return err
		}
	}
	dir, err := selectInitDir(root, home, o, bufio.NewReader(os.Stdin))
	if err != nil {
		return err
	}
	agent, agentSource, err := resolveInitAgent(filepath.Join(dir, "agent"), o.agent, o.force)
	if err != nil {
		return err
	}
	model, err := resolveInitModel(filepath.Join(dir, "model"), o.model, o.force)
	if err != nil {
		return err
	}
	prompt := o.prompt
	promptSource := "CLI --prompt"
	if o.promptFile != "" {
		b, err := readPromptFile(o.promptFile)
		if err != nil {
			return err
		}
		prompt = string(b)
		promptSource = "CLI --prompt-file: " + o.promptFile
	}
	if prompt == "" {
		if agent == "claude" {
			prompt = claudeDefaultPrompt(o.command)
			promptSource = "built-in Claude command"
		} else {
			prompt = defaultPortablePrompt()
			promptSource = "built-in portable prompt"
		}
	}
	plan := initConfigPlan{
		dir:          dir,
		agent:        agent,
		agentSource:  agentSource,
		model:        model,
		prompt:       prompt,
		promptSource: promptSource,
		force:        o.force,
	}
	if o.dryRun {
		return plan.printDryRun()
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if err := plan.apply(); err != nil {
		return err
	}
	fmt.Printf("Wrote start-issue configuration: %s\n", dir)
	return nil
}

func selectInitDir(root, home string, o options, reader *bufio.Reader) (string, error) {
	if o.project {
		if root == "" {
			return "", errors.New("--project requires a git repository")
		}
		return filepath.Join(root, ".start-issue"), nil
	}
	userDir := filepath.Join(home, ".config", "start-issue")
	if o.user {
		return userDir, nil
	}
	if root == "" {
		fmt.Printf("Initialize start-issue configuration:\n  1) User config (%s)\nChoice [1]: ", userDir)
		choice, err := reader.ReadString('\n')
		if err != nil {
			return "", errors.New("No init scope selected. Use --user outside a git repository.")
		}
		switch strings.TrimSpace(strings.ToLower(choice)) {
		case "", "1", "u", "user":
			return userDir, nil
		default:
			return "", errors.New("Project config requires a git repository. Use --user outside a git repository.")
		}
	}

	projectDir := filepath.Join(root, ".start-issue")
	fmt.Printf("Initialize start-issue configuration:\n  1) Project config (%s)\n  2) User config (%s)\nChoice [1/2]: ", projectDir, userDir)
	choice, err := reader.ReadString('\n')
	if err != nil {
		return "", errors.New("No init scope selected. Use --project or --user.")
	}
	switch strings.TrimSpace(strings.ToLower(choice)) {
	case "1", "p", "project":
		return projectDir, nil
	case "2", "u", "user":
		return userDir, nil
	default:
		return "", fmt.Errorf("Invalid init scope: %s. Use --project or --user.", strings.TrimSpace(choice))
	}
}

type initConfigPlan struct {
	dir, agent, agentSource, model, prompt, promptSource string
	force                                                bool
}

func resolveInitAgent(path, cli string, force bool) (string, string, error) {
	if !force {
		agent, err := configAgent(path, "")
		if err != nil {
			return "", "", err
		}
		if agent != "" {
			return agent, path + " (existing)", nil
		}
	}
	if cli != "" {
		if err := validateAgent(cli); err != nil {
			return "", "", err
		}
		return cli, "CLI", nil
	}
	return "claude", "built-in default", nil
}

func resolveInitModel(path, cli string, force bool) (string, error) {
	if !force {
		model, err := configValue(path, "model")
		if err != nil {
			return "", err
		}
		if model != "" {
			return model, nil
		}
	}
	if cli == "" {
		return "", nil
	}
	model := strings.TrimSpace(cli)
	if model == "" {
		return "", errors.New("Model config is empty. Remove the empty model config or set a value.")
	}
	return model, nil
}

func (p initConfigPlan) paths() (agent, model, prompt string) {
	return filepath.Join(p.dir, "agent"), filepath.Join(p.dir, "model"), filepath.Join(p.dir, "prompt.md")
}

func (p initConfigPlan) printDryRun() error {
	agentPath, modelPath, promptPath := p.paths()
	fmt.Printf("[DRY-RUN] Would create configuration in: %s\n", p.dir)
	fmt.Printf("[DRY-RUN] Agent: %s (%s)\n", p.agent, p.agentSource)
	fmt.Printf("[DRY-RUN] Prompt source: %s\n", p.promptSource)
	if err := printPlannedWrite("agent config", agentPath, p.force); err != nil {
		return err
	}
	if p.model != "" {
		if err := printPlannedWrite("model config", modelPath, p.force); err != nil {
			return err
		}
	} else if p.force {
		exists, err := pathExists(modelPath)
		if err != nil {
			return err
		}
		if exists {
			fmt.Printf("[DRY-RUN] Would remove model config: %s\n", modelPath)
		} else {
			fmt.Printf("[DRY-RUN] No model config to write (built-in default: unset)\n")
		}
	}
	return printPlannedWrite("prompt template", promptPath, p.force)
}

func printPlannedWrite(label, path string, force bool) error {
	exists, err := pathExists(path)
	if err != nil {
		return err
	}
	if exists && !force {
		fmt.Printf("[DRY-RUN] %s already exists, keeping: %s\n", label, path)
		return nil
	}
	fmt.Printf("[DRY-RUN] Would write %s: %s\n", label, path)
	return nil
}

func (p initConfigPlan) apply() error {
	agentPath, modelPath, promptPath := p.paths()
	if err := writeConfig(agentPath, p.agent+"\n", p.force); err != nil {
		return err
	}
	if p.model != "" {
		if err := writeConfig(modelPath, p.model+"\n", p.force); err != nil {
			return err
		}
	} else if p.force {
		if err := os.Remove(modelPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return writeConfig(promptPath, p.prompt+"\n", p.force)
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func setupMode(home string, dryRun bool, command string, reader *bufio.Reader) error {
	dir := filepath.Join(home, ".config", "start-issue")
	if !dryRun {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	fmt.Println("Select default agent: 1) claude  2) codex  3) kimi  4) pi  5) skip")
	fmt.Print("Choice: ")
	choice, err := reader.ReadString('\n')
	if err != nil {
		return errors.New("No setup agent selected.")
	}
	choice = strings.TrimSpace(choice)
	normalizedChoice := strings.ToLower(choice)
	agents := map[string]string{"1": "claude", "claude": "claude", "2": "codex", "codex": "codex", "3": "kimi", "kimi": "kimi", "4": "pi", "pi": "pi"}
	agent, selected := agents[normalizedChoice]
	skipAgent := normalizedChoice == "" || normalizedChoice == "5" || normalizedChoice == "skip"
	if !selected && !skipAgent {
		return errors.New("invalid setup choice")
	}
	if skipAgent {
		agent = "claude"
	}
	prompt := defaultSetupPrompt(agent, command)
	fmt.Printf("\nDefault prompt preview:\n%s\n\n", prompt)
	fmt.Print("Save a default prompt? [Y/n] ")
	answer, err := reader.ReadString('\n')
	if err != nil {
		return errors.New("No response received.")
	}
	savePrompt := false
	switch strings.TrimSpace(strings.ToLower(answer)) {
	case "", "y", "yes":
		savePrompt = true
	case "n", "no":
	default:
		return fmt.Errorf("Invalid response: %s. Use y or n.", strings.TrimSpace(answer))
	}
	if dryRun {
		fmt.Printf("[DRY-RUN] Would create configuration in: %s\n", dir)
		if savePrompt {
			fmt.Printf("[DRY-RUN] Would write prompt template: %s\n", filepath.Join(dir, "prompt.md"))
		} else {
			fmt.Printf("[DRY-RUN] Would remove prompt template: %s\n", filepath.Join(dir, "prompt.md"))
		}
		if skipAgent {
			fmt.Printf("[DRY-RUN] Would remove agent config: %s\n", filepath.Join(dir, "agent"))
		} else {
			fmt.Printf("[DRY-RUN] Would write agent config: %s\n", filepath.Join(dir, "agent"))
		}
		return nil
	}
	if savePrompt {
		if err := writeConfig(filepath.Join(dir, "prompt.md"), prompt+"\n", true); err != nil {
			return err
		}
	} else if err := removeConfigIfExists(filepath.Join(dir, "prompt.md")); err != nil {
		return fmt.Errorf("remove prompt template: %w", err)
	}
	if skipAgent {
		if err := removeConfigIfExists(filepath.Join(dir, "agent")); err != nil {
			return fmt.Errorf("remove agent config: %w", err)
		}
	} else if err := writeConfig(filepath.Join(dir, "agent"), agent+"\n", true); err != nil {
		return err
	}
	fmt.Printf("Wrote start-issue configuration: %s\n", dir)
	return nil
}

func removeConfigIfExists(path string) error {
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func defaultSetupPrompt(agent, command string) string {
	if agent == "claude" {
		return claudeDefaultPrompt(command)
	}
	return defaultPortablePrompt()
}

func installMode(dryRun bool) error {
	name, err := releaseAssetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		return errors.New("Windows installation is manual: download start-issue-windows-amd64.exe from the latest release")
	}
	home, err := userHomeDir()
	if err != nil {
		return err
	}
	target := filepath.Join(home, ".local", "bin", "start-issue")
	if dryRun {
		fmt.Printf("[DRY-RUN] Would download %s, verify checksums.txt, and install: %s\n", name, target)
		return nil
	}
	if err := checkGitHubAccess(); err != nil {
		return err
	}
	data, err := output("gh", "api", "repos/"+releaseRepository()+"/releases/latest")
	if err != nil {
		return errors.New("Could not fetch the latest start-issue release")
	}
	var release githubRelease
	if err := json.Unmarshal([]byte(data), &release); err != nil {
		return err
	}
	assetURL, checksumURL := release.assetURLs(name)
	if assetURL == "" || checksumURL == "" {
		return fmt.Errorf("latest release does not contain %s and checksums.txt", name)
	}
	binary, err := download(assetURL)
	if err != nil {
		return err
	}
	checksums, err := download(checksumURL)
	if err != nil {
		return err
	}
	if !validChecksum(binary, name, string(checksums)) {
		return fmt.Errorf("checksum verification failed for %s", name)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	if err := installVerifiedUpdate(target, binary, release.TagName); err != nil {
		return err
	}
	fmt.Printf("Installed start-issue v%s at: %s\n", strings.TrimPrefix(release.TagName, "v"), target)
	return nil
}

func installBinary(target string, binary []byte) error {
	temporary, err := stageBinary(target, binary)
	if err != nil {
		return err
	}
	defer os.Remove(temporary)
	if err := os.Rename(temporary, target); err != nil {
		return err
	}
	return nil
}

func writeConfig(path, content string, force bool) error {
	if _, err := os.Stat(path); err == nil && !force {
		return nil
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func configAgent(path, fallback string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fallback, nil
		}
		return "", err
	}
	agent := first(string(b))
	if agent == "" {
		return "", errors.New("Agent config is empty. Remove the empty agent config or set a value.")
	}
	switch agent {
	case "claude", "codex", "kimi", "pi", "none":
		return agent, nil
	default:
		return "", fmt.Errorf("Unknown agent: %s. Valid agents: claude, codex, kimi, pi, none.", agent)
	}
}

func configValue(path, name string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	value := first(string(b))
	if value == "" {
		return "", fmt.Errorf("%s config is empty. Remove the empty %s config or set a value.", strings.ToUpper(name[:1])+name[1:], name)
	}
	return value, nil
}

func updateMode(o options) error {
	assetName, err := releaseAssetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		fmt.Println("Windows update is manual: download start-issue-windows-amd64.exe from the latest release and replace the executable on PATH.")
		return nil
	}
	if err := checkGitHubAccess(); err != nil {
		return err
	}
	data, err := output("gh", "api", "repos/"+releaseRepository()+"/releases/latest")
	if err != nil {
		return errors.New("Could not fetch the latest start-issue release")
	}
	var release githubRelease
	if err := json.Unmarshal([]byte(data), &release); err != nil {
		return fmt.Errorf("decode latest release: %w", err)
	}
	if strings.TrimSpace(release.TagName) == "" {
		return errors.New("latest release response is missing tag_name")
	}
	if compareVersions(runningVersion(), release.TagName) >= 0 {
		fmt.Printf("start-issue is already up to date (%s).\n", runningVersion())
		return nil
	}
	assetURL, checksumURL := release.assetURLs(assetName)
	if assetURL == "" || checksumURL == "" {
		return fmt.Errorf("latest release does not contain %s and checksums.txt", assetName)
	}
	if o.dryRun {
		target, err := runningExecutablePath()
		if err != nil {
			return err
		}
		fmt.Printf("[DRY-RUN] Would download %s, verify checksums.txt, and replace: %s\n", assetName, target)
		return nil
	}
	binary, err := download(assetURL)
	if err != nil {
		return fmt.Errorf("download %s: %w", assetName, err)
	}
	checksums, err := download(checksumURL)
	if err != nil {
		return fmt.Errorf("download checksums.txt: %w", err)
	}
	if !validChecksum(binary, assetName, string(checksums)) {
		return fmt.Errorf("checksum verification failed for %s", assetName)
	}
	target, err := runningExecutablePath()
	if err != nil {
		return err
	}
	if err := installVerifiedUpdate(target, binary, release.TagName); err != nil {
		return err
	}
	fmt.Printf("Updated start-issue at: %s\nVersion: start-issue v%s\n", target, strings.TrimPrefix(release.TagName, "v"))
	return nil
}

func runningExecutablePath() (string, error) {
	target, err := os.Executable()
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(target)
	if err != nil {
		return "", fmt.Errorf("resolve running executable %q: %w", target, err)
	}
	return resolved, nil
}

func installVerifiedUpdate(target string, binary []byte, expectedTag string) error {
	temporary, err := stageBinary(target, binary)
	if err != nil {
		return err
	}
	defer os.Remove(temporary)
	if err := verifyStagedBinary(temporary, expectedTag); err != nil {
		return err
	}
	if err := os.Rename(temporary, target); err != nil {
		return err
	}
	return nil
}

func stageBinary(target string, binary []byte) (string, error) {
	temporary, err := os.CreateTemp(filepath.Dir(target), "."+filepath.Base(target)+".new-*")
	if err != nil {
		return "", err
	}
	path := temporary.Name()
	cleanup := func(err error) (string, error) {
		_ = temporary.Close()
		_ = os.Remove(path)
		return "", err
	}
	if err := temporary.Chmod(0755); err != nil {
		return cleanup(err)
	}
	if _, err := temporary.Write(binary); err != nil {
		return cleanup(err)
	}
	if err := temporary.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	return path, nil
}

func verifyStagedBinary(path, expectedTag string) error {
	result, err := exec.Command(path, "--version").Output()
	if err != nil {
		return fmt.Errorf("staged update failed version verification: %w", err)
	}
	got := strings.TrimSpace(string(result))
	want := "start-issue v" + strings.TrimPrefix(expectedTag, "v")
	if got != want {
		return fmt.Errorf("staged update version %q does not match expected release %q", got, expectedTag)
	}
	return nil
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

func releaseRepository() string {
	if repository := strings.TrimSpace(os.Getenv("START_ISSUE_REPOSITORY")); repository != "" {
		return repository
	}
	return defaultReleaseRepository
}

func (r githubRelease) assetURLs(name string) (string, string) {
	var assetURL, checksumURL string
	for _, asset := range r.Assets {
		switch asset.Name {
		case name:
			assetURL = asset.URL
		case "checksums.txt":
			checksumURL = asset.URL
		}
	}
	return assetURL, checksumURL
}

func releaseAssetName(goos, goarch string) (string, error) {
	switch goos + "/" + goarch {
	case "linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64":
		return fmt.Sprintf("start-issue-%s-%s", goos, goarch), nil
	case "windows/amd64":
		return "start-issue-windows-amd64.exe", nil
	default:
		return "", fmt.Errorf("unsupported release platform %s/%s; supported platforms: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64", goos, goarch)
	}
}

func download(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("unexpected HTTP status %s", response.Status)
	}
	return io.ReadAll(response.Body)
}

func validChecksum(binary []byte, name, manifest string) bool {
	want := ""
	for _, line := range strings.Split(manifest, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && strings.TrimPrefix(fields[1], "*") == name {
			want = fields[0]
			break
		}
	}
	actual := fmt.Sprintf("%x", sha256.Sum256(binary))
	return want != "" && strings.EqualFold(want, actual)
}

type semanticVersion struct {
	core       [3]int
	prerelease []string
	postTag    bool
}

func parseSemanticVersion(value string) semanticVersion {
	value = strings.TrimPrefix(strings.TrimSpace(value), "v")
	if build := strings.IndexByte(value, '+'); build >= 0 {
		value = value[:build]
	}
	// Make builds use git describe. Its "-<distance>-g<commit>" suffix, or
	// "-dirty" when the checkout is modified exactly at a tag, means this
	// source build is newer than the referenced tag, not a SemVer prerelease.
	// Preserve that distinction so update never downgrades it.
	postTag := false
	if matches := gitDescribeSuffix.FindStringSubmatch(value); matches != nil {
		value = matches[1]
		postTag = true
	}
	parts := strings.SplitN(value, "-", 2)
	core := strings.Split(parts[0], ".")
	result := semanticVersion{postTag: postTag}
	for index := range result.core {
		if index < len(core) {
			result.core[index], _ = strconv.Atoi(core[index])
		}
	}
	if len(parts) == 2 && parts[1] != "" {
		result.prerelease = strings.Split(parts[1], ".")
	}
	return result
}

func compareVersions(left, right string) int {
	a, b := parseSemanticVersion(left), parseSemanticVersion(right)
	for index := range a.core {
		if a.core[index] < b.core[index] {
			return -1
		}
		if a.core[index] > b.core[index] {
			return 1
		}
	}
	if len(a.prerelease) == 0 && len(b.prerelease) > 0 {
		return 1
	}
	if len(a.prerelease) > 0 && len(b.prerelease) == 0 {
		return -1
	}
	for index := 0; index < len(a.prerelease) && index < len(b.prerelease); index++ {
		if result := comparePrereleaseIdentifier(a.prerelease[index], b.prerelease[index]); result != 0 {
			return result
		}
	}
	if len(a.prerelease) < len(b.prerelease) {
		return -1
	}
	if len(a.prerelease) > len(b.prerelease) {
		return 1
	}
	if a.postTag && !b.postTag {
		return 1
	}
	if !a.postTag && b.postTag {
		return -1
	}
	return 0
}

func comparePrereleaseIdentifier(left, right string) int {
	leftNumber, leftErr := strconv.Atoi(left)
	rightNumber, rightErr := strconv.Atoi(right)
	if leftErr == nil && rightErr == nil {
		return compareInts(leftNumber, rightNumber)
	}
	if leftErr == nil {
		return -1
	}
	if rightErr == nil {
		return 1
	}
	return strings.Compare(left, right)
}

func compareInts(left, right int) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}

func resolveAgent(root, cli string) (string, string, error) {
	project := ""
	if root != "" {
		project = filepath.Join(root, ".start-issue", "agent")
	}
	v, s, e := resolve(cli, project, userConfigPath("agent"), "START_ISSUE_AGENT", "claude")
	if e != nil {
		return "", "", e
	}
	if v == "" {
		return "", s, errors.New("Agent config is empty. Valid agents: claude, codex, kimi, pi, none.")
	}
	if err := validateAgent(v); err != nil {
		return "", "", err
	}
	return v, s, nil
}

func validateAgent(agent string) error {
	switch agent {
	case "claude", "codex", "kimi", "pi", "none":
		return nil
	default:
		return fmt.Errorf("Unknown agent: %s. Valid agents: claude, codex, kimi, pi, none.", agent)
	}
}
func resolveModel(root, cli string) (string, string, error) {
	if cli != "" {
		cli = strings.TrimSpace(cli)
		if cli == "" {
			return "", "CLI", errors.New("--model requires a non-empty value.")
		}
	}
	project := ""
	if root != "" {
		project = filepath.Join(root, ".start-issue", "model")
	}
	v, s, e := resolve(cli, project, userConfigPath("model"), "START_ISSUE_MODEL", "")
	if e == nil && s != "built-in default" && v == "" {
		e = errors.New("Model config is empty. Remove the empty model config or set a value.")
	}
	return v, s, e
}
func resolve(cli, project, user, env, def string) (string, string, error) {
	if cli != "" {
		return cli, "CLI", nil
	}
	for _, p := range []string{project, user} {
		if p == "" {
			continue
		}
		if b, e := os.ReadFile(p); e == nil {
			return first(string(b)), p, nil
		} else if !os.IsNotExist(e) {
			return "", "", e
		}
	}
	if raw, ok := os.LookupEnv(env); ok && raw != "" {
		return strings.TrimSpace(raw), env, nil
	}
	return def, "built-in default", nil
}
func first(s string) string {
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(strings.SplitN(l, "#", 2)[0]); l != "" {
			return l
		}
	}
	return ""
}

func userConfigPath(name string) string {
	home, err := userHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "start-issue", name)
}

// resolvePrompt returns the rendered template, its display source and location,
// and (only for file-backed prompts) the resolved source file path.
func resolvePrompt(root, agent string, o options) (string, string, string, string, error) {
	if o.prompt != "" {
		return o.prompt, "CLI --prompt", "inline CLI argument", "", nil
	}
	if o.promptFile != "" {
		b, e := readPromptFile(o.promptFile)
		location, locationErr := absolutePath(o.promptFile)
		if e != nil {
			return "", "", "", "", e
		}
		return string(b), "CLI --prompt-file: " + o.promptFile, location, location, locationErr
	}
	if os.Getenv("START_ISSUE_PROMPT_FILE") != "" && os.Getenv("START_ISSUE_PROMPT") != "" {
		return "", "", "", "", errors.New("Use either START_ISSUE_PROMPT_FILE or START_ISSUE_PROMPT, not both.")
	}
	if path := os.Getenv("START_ISSUE_PROMPT_FILE"); path != "" {
		b, e := readPromptFile(path)
		location, locationErr := absolutePath(path)
		if e != nil {
			return "", "", "", "", e
		}
		return string(b), "START_ISSUE_PROMPT_FILE: " + path, location, location, locationErr
	}
	if value := os.Getenv("START_ISSUE_PROMPT"); value != "" {
		return value, "START_ISSUE_PROMPT", "START_ISSUE_PROMPT environment variable", "", nil
	}
	paths := []string{}
	if userConfig := userConfigPath("prompt.md"); userConfig != "" {
		paths = append(paths, userConfig)
	}
	if root != "" {
		paths = append([]string{filepath.Join(root, ".start-issue", "prompt.md")}, paths...)
	}
	for _, path := range paths {
		if b, e := readPromptFile(path); e == nil {
			return string(b), path, path, path, nil
		} else if !errors.Is(e, os.ErrNotExist) {
			return "", "", "", "", e
		}
	}
	location, err := os.Executable()
	if err != nil {
		return "", "", "", "", err
	}
	if agent == "claude" {
		return claudeDefaultPrompt(o.command), "built-in Claude command", location, "", nil
	}
	return defaultPortablePrompt(), "built-in portable prompt", location, "", nil
}

func readPromptFile(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, promptFileNotFoundError{path: path, cause: err}
	}
	// Bash command substitution, used by the v1 implementation to read prompt
	// files, removes every trailing newline. Keep file-backed prompt sources
	// compatible so init subsequently writes exactly one trailing newline.
	return bytes.TrimRight(b, "\n"), err
}

func absolutePath(path string) (string, error) { return filepath.Abs(path) }
func claudeDefaultPrompt(command string) string {
	if command != "" {
		return command + " {ISSUE_URL}"
	}
	return "/task-router:route-task {ISSUE_URL}"
}
func defaultPortablePrompt() string {
	return `Implement GitHub issue {ISSUE_URL} in this worktree.

Context:
- Repo: {REPO}
- Issue: #{ISSUE_NUMBER}
- Title: {ISSUE_TITLE}
- Branch: {BRANCH_NAME}
- Worktree: {WORKTREE_PATH}

Start by reading the issue with gh if needed. Follow repository instructions. Keep changes scoped. Run relevant tests or checks. Summarize changed files and verification before finishing.
If you open a PR for this work, target the base branch {BASE_BRANCH}.`
}
func parseIssue(v, repo string) (string, string, error) {
	re := regexp.MustCompile(`^https://github\.com/([^/]+)/([^/]+)/issues/([0-9]+)`)
	if m := re.FindStringSubmatch(v); m != nil {
		return m[3], m[1] + "/" + m[2], nil
	}
	if regexp.MustCompile(`^[0-9]+$`).MatchString(v) {
		return v, repo, nil
	}
	return "", "", fmt.Errorf("Invalid issue format: %s. Use issue number or full GitHub URL.", v)
}
func detectRepo() (string, error) {
	v, e := output("git", "remote", "get-url", "origin")
	if e != nil {
		return "", errors.New("Cannot detect repository. No 'origin' remote found. Use --repo flag.")
	}
	v = strings.TrimSuffix(strings.TrimSpace(v), ".git")
	return repoFromRemoteURL(v)
}

func repoFromRemoteURL(v string) (string, error) {
	for _, p := range []string{"git@github.com:", "https://github.com/"} {
		if strings.HasPrefix(v, p) {
			return strings.TrimPrefix(v, p), nil
		}
	}
	return "", fmt.Errorf("Cannot parse repository from remote URL: %s. Use --repo flag.", v)
}
func detectBase() string {
	if v, e := output("git", "symbolic-ref", "refs/remotes/origin/HEAD"); e == nil {
		return strings.TrimPrefix(strings.TrimSpace(v), "refs/remotes/origin/")
	}
	v, _ := output("git", "rev-parse", "--abbrev-ref", "HEAD")
	return strings.TrimSpace(v)
}
func branchName(n, title, labels string) string {
	kind := "feature"
	switch {
	case containsAny(labels, "hotfix", "critical", "urgent"):
		kind = "hotfix"
	case containsAny(labels, "bug", "fix", "bugfix", "error"):
		kind = "fix"
	case containsAny(labels, "docs", "documentation"):
		kind = "docs"
	case containsAny(labels, "refactor", "tech-debt", "cleanup", "technical"):
		kind = "refactor"
	case containsAny(labels, "test", "testing", "tests"):
		kind = "test"
	case containsAny(labels, "chore", "ci", "build", "infra"):
		kind = "chore"
	}
	slug := slugify(title)
	return fmt.Sprintf("%s/issue-%s-%s", kind, n, slug)
}

func containsAny(s string, terms ...string) bool {
	for _, term := range terms {
		if strings.Contains(s, term) {
			return true
		}
	}
	return false
}
func aiBranchName(agent, model, root, number, title, labels string) (string, error) {
	if agent == "none" {
		return "", errors.New("no agent")
	}
	prompt := fmt.Sprintf("Git branch name for issue #%s: %q (labels: %s).\nFormat: {type}/issue-%s-{kebab-case-name}\nTypes: bug/fix -> fix, enhancement -> feature, hotfix -> hotfix, docs -> docs, refactor -> refactor, test -> test, chore -> chore, default -> feature.\nIf the title contains non-English text (for example Cyrillic), transliterate it to English for the kebab-case name.\nStrip leading bracketed process/stage tags (for example [brief], [investigation], [PR-008]) from the kebab-case name; they mark workflow stage. The {type} still comes from the labels above.\nReply with ONLY the branch name.", number, title, labels, number)
	args := helperArgs(agent, model, root, prompt)
	result, err := helperOutput(agent, root, args)
	if err != nil {
		return "", err
	}
	var branch string
	for _, line := range strings.Split(result, "\n") {
		line = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(line, "`", ""), "\"", ""))
		if line != "" {
			branch = line
		}
	}
	if branch == "" {
		return "", errors.New("empty branch")
	}
	return branch, nil
}

func slugify(title string) string {
	title = regexp.MustCompile(`^(\[[^]]*\][\s-]*)+`).ReplaceAllString(title, "")
	var b strings.Builder
	for _, r := range strings.ToLower(title) {
		if replacement, ok := cyrillicTransliteration[r]; ok {
			b.WriteString(replacement)
			continue
		}
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		} else {
			b.WriteByte('-')
		}
	}
	slug := regexp.MustCompile(`-+`).ReplaceAllString(b.String(), "-")
	if len(slug) > 40 {
		slug = slug[:40]
	}
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "work"
	}
	return slug
}

var cyrillicTransliteration = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "yo", 'ж': "zh", 'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m", 'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u", 'ф': "f", 'х': "kh", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "shch", 'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
}

func render(s string, m map[string]string) string {
	// Preserve the Bash implementation's ordered substitutions. In particular,
	// placeholders in issue metadata are expanded when their replacement occurs
	// later in this sequence.
	for _, key := range []string{
		"ISSUE_URL",
		"ISSUE_NUMBER",
		"ISSUE_TITLE",
		"ISSUE_BODY",
		"ISSUE_LABELS",
		"REPO",
		"BRANCH_NAME",
		"WORKTREE_PATH",
		"BASE_BRANCH",
	} {
		if value, ok := m[key]; ok {
			s = strings.ReplaceAll(s, "{"+key+"}", value)
		}
	}
	return s
}

func renderIssuePrompt(prompt, issueURL, number string, in issue, labels []string, repo, branch, worktree, base string) string {
	return render(prompt, map[string]string{
		"ISSUE_URL":     issueURL,
		"ISSUE_NUMBER":  number,
		"ISSUE_TITLE":   in.Title,
		"ISSUE_BODY":    in.Body,
		"ISSUE_LABELS":  strings.Join(labels, ", "),
		"REPO":          repo,
		"BRANCH_NAME":   branch,
		"WORKTREE_PATH": worktree,
		"BASE_BRANCH":   base,
	})
}

func worktreePath(parent, branch string, flat bool) string {
	name := branch
	if flat {
		name = strings.ReplaceAll(name, "/", "-")
	}
	return filepath.Join(parent, name)
}

func branchRefExists(branch string) bool {
	_, err := output("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	return err == nil
}

func nextSuffixedBranch(branch string) string {
	for version := 2; ; version++ {
		candidate := fmt.Sprintf("%s-v%d", branch, version)
		if !branchRefExists(candidate) {
			return candidate
		}
	}
}

func branchConflictChoice(branch, existing string, reader *bufio.Reader) string {
	fmt.Printf("\nBranch '%s' already exists.\n", branch)
	if existing != "" {
		fmt.Printf("   Existing worktree: %s\n", existing)
	}
	fmt.Print("\n  1) Use existing worktree and continue\n  2) Create new branch with different name\n  3) Delete branch/worktree and recreate\n  0) Exit\n\nChoice: ")
	return readChoice(reader)
}

func pathConflictChoice(worktree, branch string, registered bool, reader *bufio.Reader) (string, error) {
	fmt.Printf("\nWorktree path already exists: %s\n", worktree)
	if !registered {
		return "", fmt.Errorf("Cannot use worktree path '%s': it exists but is not a registered git worktree. Move or remove the directory manually, then rerun.", worktree)
	}
	if branch == "" {
		fmt.Println("   Registered branch: detached HEAD")
	} else {
		fmt.Printf("   Registered branch: %s\n", strings.TrimPrefix(branch, "refs/heads/"))
	}
	fmt.Print("\n  1) Use existing worktree\n  2) Delete and recreate\n  0) Exit\n\nChoice: ")
	return readChoice(reader), nil
}

func readChoice(reader *bufio.Reader) string {
	answer, _ := reader.ReadString('\n')
	return strings.TrimSpace(answer)
}

func validateReusedWorktree(path, branch string) error {
	if !fileExists(path) {
		return fmt.Errorf("Cannot reuse worktree path '%s': directory does not exist.", path)
	}
	registered, found := worktreeRegistration(path)
	if !found {
		return fmt.Errorf("Cannot reuse worktree path '%s': path exists but is not a git worktree for this repository.", path)
	}
	if registered == "" {
		return fmt.Errorf("Cannot reuse worktree path '%s': it is a detached git worktree; expected branch '%s'.", path, branch)
	}
	if registered != "refs/heads/"+branch {
		return fmt.Errorf("Cannot reuse worktree path '%s': it belongs to branch '%s', not '%s'.", path, strings.TrimPrefix(registered, "refs/heads/"), branch)
	}
	return nil
}

func removeWorktreeAndBranch(worktree, branch string) error {
	if worktree != "" {
		if fileExists(worktree) {
			removable, err := removableLinkedWorktree(worktree)
			if err != nil {
				return fmt.Errorf("could not verify that worktree path %q is removable: %w", worktree, err)
			}
			if !removable {
				return fmt.Errorf("refusing to remove %q: it is not a removable linked worktree", worktree)
			}
		}
		if err := command("git", "worktree", "remove", "--force", worktree); err != nil {
			if fileExists(worktree) {
				return fmt.Errorf("refusing to delete worktree path %q after git worktree remove failed: %w", worktree, err)
			}
			if err := command("git", "worktree", "prune"); err != nil {
				return err
			}
		}
	}
	if branch != "" {
		if err := command("git", "branch", "-D", branch); err != nil {
			return err
		}
	}
	return nil
}

// removableLinkedWorktree reports whether path is a registered non-primary,
// non-current worktree. Git's primary and current worktrees must never be
// removed by this command.
func removableLinkedWorktree(path string) (bool, error) {
	list, err := output("git", "worktree", "list", "--porcelain")
	if err != nil {
		return false, err
	}
	target := canonicalPath(path)
	current, err := output("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return false, err
	}
	if canonicalPath(strings.TrimSpace(current)) == target {
		return false, nil
	}
	worktreeIndex := -1
	for _, line := range strings.Split(list, "\n") {
		if !strings.HasPrefix(line, "worktree ") {
			continue
		}
		worktreeIndex++
		if canonicalPath(strings.TrimPrefix(line, "worktree ")) == target {
			return worktreeIndex > 0, nil
		}
	}
	return false, nil
}
func branchWorktree(branch string) string {
	value, err := output("git", "worktree", "list", "--porcelain")
	if err != nil {
		return ""
	}
	current := ""
	for _, line := range strings.Split(value, "\n") {
		if strings.HasPrefix(line, "worktree ") {
			current = strings.TrimPrefix(line, "worktree ")
		}
		if line == "branch refs/heads/"+branch {
			return current
		}
	}
	return ""
}
func worktreeBranch(path string) string {
	branch, _ := worktreeRegistration(path)
	return branch
}

// worktreeRegistration distinguishes a registered detached worktree from a
// plain directory. Detached worktrees do not have a branch record in Git's
// porcelain output, but can still be safely removed through Git.
func worktreeRegistration(path string) (string, bool) {
	path = canonicalPath(path)
	value, err := output("git", "worktree", "list", "--porcelain")
	if err != nil {
		return "", false
	}
	current := ""
	registered := false
	currentMatches := false
	branch := ""
	for _, line := range strings.Split(value, "\n") {
		if strings.HasPrefix(line, "worktree ") {
			current = strings.TrimPrefix(line, "worktree ")
			currentMatches = canonicalPath(current) == path
			registered = registered || currentMatches
		}
		if currentMatches && strings.HasPrefix(line, "branch ") {
			branch = strings.TrimPrefix(line, "branch ")
		}
	}
	return branch, registered
}
func canonicalPath(path string) string {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	if resolved, err := filepath.EvalSymlinks(absolute); err == nil {
		return resolved
	}
	return absolute
}
func launchSelected(o options, agent, model, worktree, prompt string) error {
	if o.dryRun {
		if o.humanGate {
			return humanGate(model, worktree, prompt, true)
		}
		if agent == "none" {
			printManualNextSteps(model, worktree)
			return nil
		}
		printLaunch(agent, model, worktree, prompt)
		return nil
	}
	if o.humanGate {
		return humanGate(model, worktree, prompt, false)
	}
	return launch(agent, model, worktree, prompt)
}
func improvePrompt(root, agent, model, prompt, source, promptFile string, o options, in issue, repo, number, labels string) error {
	if agent == "none" {
		return errors.New("--improve-prompt requires an agent. Use --agent claude, codex, kimi, or pi.")
	}
	outputPath := promptImprovementOutputPath(root, promptFile, o)
	if o.dryRun {
		fmt.Printf("📝 Improving prompt template...\n   Prompt source: %s\n   Proposal path: %s\n   [DRY-RUN] Would ask %s to generate an improved prompt proposal.\n", source, outputPath, agent)
		return nil
	}
	if _, err := os.Stat(outputPath); err == nil {
		return fmt.Errorf("Prompt improvement output already exists: %s", outputPath)
	}
	request := promptImprovementRequest(prompt, source, in, repo, number, labels)
	args := helperArgs(agent, model, root, request)
	result, err := helperOutput(agent, root, args)
	if err != nil {
		return fmt.Errorf("Could not generate improved prompt with %s", agent)
	}
	proposal := normalizePromptProposal(result)
	if proposal == "" {
		return errors.New("Improved prompt proposal is empty.")
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, []byte(proposal+"\n"), 0644); err != nil {
		return err
	}
	fmt.Printf("📝 Prompt improvement written: %s\n", outputPath)
	return nil
}

func promptImprovementOutputPath(root, promptFile string, o options) string {
	if o.promptOutput != "" {
		return o.promptOutput
	}
	if promptFile == "" {
		return filepath.Join(root, ".start-issue", "prompt.improved.md")
	}
	if filepath.Ext(promptFile) == ".md" {
		return strings.TrimSuffix(promptFile, ".md") + ".improved.md"
	}
	return promptFile + ".improved"
}

func promptImprovementRequest(prompt, source string, in issue, repo, number, labels string) string {
	issueURL := fmt.Sprintf("https://github.com/%s/issues/%s", repo, number)
	return fmt.Sprintf("Improve the following start-issue prompt template.\n\nReturn ONLY the complete improved prompt template. Do not include commentary, code fences, diffs, or explanations.\n\nPreserve any placeholders that are still useful. Supported placeholders:\n{ISSUE_URL}, {ISSUE_NUMBER}, {ISSUE_TITLE}, {ISSUE_BODY}, {ISSUE_LABELS}, {REPO}, {BRANCH_NAME}, {WORKTREE_PATH}, {BASE_BRANCH}\n\nPrompt source:\n%s\n\nRepository:\n%s\n\nCurrent issue used as improvement context:\n- URL: %s\n- Number: %s\n- Title: %s\n- Labels: %s\n- Body:\n%s\n\nCurrent prompt template:\n--- START PROMPT TEMPLATE ---\n%s\n--- END PROMPT TEMPLATE ---", source, repo, issueURL, number, in.Title, labels, in.Body, prompt)
}

func normalizePromptProposal(result string) string {
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) >= 2 && regexp.MustCompile("^```[[:alnum:]_-]*\\s*$").MatchString(strings.TrimSpace(lines[0])) && strings.TrimSpace(lines[len(lines)-1]) == "```" {
		lines = lines[1 : len(lines)-1]
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
func humanGate(model, worktree, prompt string, dryRun bool) error {
	runID := os.Getenv("START_ISSUE_RUN_ID")
	if runID == "" {
		runID = time.Now().Format("20060102-150405")
	}
	dir := filepath.Join(worktree, ".start-issue", "runs", runID)
	events, last := filepath.Join(dir, "events.jsonl"), filepath.Join(dir, "last-message.txt")
	args := []string{"exec", "--cd", worktree, "--sandbox", "workspace-write", "--json", "--output-last-message", last, "-"}
	if model != "" {
		args = append([]string{"exec", "--model", model}, args[1:]...)
	}
	if dryRun {
		threadID := filepath.Join(dir, "thread-id")
		fmt.Printf("   [DRY-RUN] Would run: codex %s > %s\n", shellJoin(args), shellQuote(events))
		fmt.Printf("   [DRY-RUN] Would write captured thread ID: %s\n", threadID)
		return nil
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	file, err := os.Create(events)
	if err != nil {
		return err
	}
	defer file.Close()
	cmd := exec.Command("codex", args...)
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Stdout = file
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
	threadID, err := captureThreadID(events)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "thread-id"), []byte(threadID+"\n"), 0644); err != nil {
		return err
	}
	body, err := os.ReadFile(last)
	if err != nil {
		return fmt.Errorf("No recognized final status found. Inspect: %s", last)
	}
	status := finalStatus(string(body))
	if status == "DONE" {
		fmt.Println("✅ Codex finished with STATUS: DONE")
		return nil
	}
	if status == "HUMAN_GATE" {
		resume := []string{"resume", "--include-non-interactive", threadID}
		fmt.Printf("Resume command: codex %s\nThread ID: %s\n", strings.Join(resume, " "), threadID)
		if err := command("codex", resume...); err != nil {
			fmt.Fprintln(os.Stderr, "Could not open Codex resume session.")
			return exitError{code: 2, err: errors.New("Could not open Codex resume session.")}
		}
		return nil
	}
	return fmt.Errorf("No recognized final status found. Inspect: %s", last)
}

func captureThreadID(events string) (string, error) {
	eventsBody, err := os.ReadFile(events)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(eventsBody), "\n") {
		var event struct {
			Type     string `json:"type"`
			ThreadID string `json:"thread_id"`
		}
		if json.Unmarshal([]byte(line), &event) == nil && event.Type == "thread.started" && event.ThreadID != "" {
			return event.ThreadID, nil
		}
	}
	return "", fmt.Errorf("Codex human-gate run did not capture thread_id. Inspect: %s", events)
}

func finalStatus(body string) string {
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "STATUS:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "STATUS:"))
		}
	}
	return ""
}
func printLaunch(a, m, w, p string) {
	if a == "none" {
		fmt.Printf("   Agent: none\n   Model: %s\n   [DRY-RUN] Would prepare worktree without launching an agent\n", show(m))
		return
	}
	promptLength := utf8.RuneCountInString(p)
	fmt.Printf("   Prompt length: %d chars\n", promptLength)
	if promptLength > 4000 && os.Getenv("START_ISSUE_DUMP_PROMPT") != "1" {
		fmt.Println("   Prompt omitted from command display because it is large.")
		fmt.Println("   Set START_ISSUE_DUMP_PROMPT=1 to print the full rendered prompt.")
		p = fmt.Sprintf("<rendered prompt: %d chars>", promptLength)
	}
	fmt.Printf("   Agent: %s\n   Model: %s\n   [DRY-RUN] Would run: %s\n", a, show(m), launchDisplayCommand(a, m, w, p))
}

func launchDisplayCommand(a, m, w, p string) string {
	command := shellJoin(launchArgs(a, m, w, p))
	if a == "claude" || a == "kimi" || a == "pi" {
		return "cd " + shellQuote(w) + " && " + command
	}
	return command
}
func launch(a, m, w, p string) error {
	if a == "none" {
		printManualNextSteps(m, w)
		return nil
	}
	args := launchArgs(a, m, w, p)
	var err error
	if a == "claude" || a == "kimi" || a == "pi" {
		err = commandAt(w, args...)
	} else {
		err = command(args[0], args[1:]...)
	}
	if err != nil {
		var exited *exec.ExitError
		if errors.As(err, &exited) {
			return exitError{code: processExitCode(exited), err: err}
		}
		return err
	}
	return nil
}

func processExitCode(exited *exec.ExitError) int {
	if status, ok := exited.Sys().(syscall.WaitStatus); ok && status.Signaled() {
		return 128 + int(status.Signal())
	}
	return exited.ExitCode()
}

func printManualNextSteps(model, worktree string) {
	fmt.Printf("✅ Worktree ready at: %s\n\n", worktree)
	fmt.Printf("Selected agent: none\nResolved model: %s\n", show(model))
	fmt.Println("To start working:")
	fmt.Printf("  cd %s\n\n", shellQuote(worktree))
	fmt.Println("Suggested agent commands:")
	fmt.Println("  claude")
	fmt.Printf("  codex --cd %s\n", shellQuote(worktree))
	fmt.Printf("  (cd %s && kimi)\n", shellQuote(worktree))
	fmt.Println("  pi")
}

func shellQuote(value string) string {
	if value != "" && !strings.ContainsAny(value, " \t\r\n'\"\\$`!()[]{}*?;<>&|#~") {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func shellJoin(values []string) string {
	quoted := make([]string, len(values))
	for i, value := range values {
		quoted[i] = shellQuote(value)
	}
	return strings.Join(quoted, " ")
}
func launchArgs(a, m, w, p string) []string {
	switch a {
	case "none":
		return nil
	case "claude":
		x := []string{"claude"}
		if m != "" {
			x = append(x, "--model", m)
		}
		return append(x, "--dangerously-skip-permissions", p)
	case "codex":
		x := []string{"codex"}
		if m != "" {
			x = append(x, "--model", m)
		}
		return append(x, "--cd", w, "--dangerously-bypass-approvals-and-sandbox", p)
	case "kimi":
		x := []string{"kimi"}
		if m != "" {
			x = append(x, "--model", m)
		}
		return append(x, "-p", p)
	default:
		x := []string{"pi"}
		if m != "" {
			x = append(x, "--model", m)
		}
		return append(x, p)
	}
}

func helperArgs(agent, model, root, prompt string) []string {
	withModel := func(args []string) []string {
		if model != "" {
			return append(args, "--model", model)
		}
		return args
	}
	switch agent {
	case "claude":
		if model == "" {
			model = "haiku"
		}
		args := withModel([]string{"claude", "--print"})
		return append(args, "--no-session-persistence", "--disable-slash-commands", prompt)
	case "codex":
		args := []string{"codex", "exec"}
		if model != "" {
			args = append(args, "--model", model)
		}
		return append(args, "--cd", root, "--sandbox", "read-only", "--skip-git-repo-check", prompt)
	case "kimi":
		args := []string{"kimi"}
		if model != "" {
			args = append(args, "--model", model)
		}
		return append(args, "-p", prompt)
	case "pi":
		args := []string{"pi"}
		if model != "" {
			args = append(args, "--model", model)
		}
		return append(args, "--print", "--no-tools", "--no-session", prompt)
	default:
		return nil
	}
}
func helperOutput(agent, root string, args []string) (string, error) {
	if agent == "kimi" {
		return outputAt(root, args[0], args[1:]...)
	}
	return output(args[0], args[1:]...)
}
func command(name string, args ...string) error {
	c := exec.Command(name, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	return c.Run()
}
func commandAt(dir string, args ...string) error {
	c := exec.Command(args[0], args[1:]...)
	c.Dir = dir
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	return c.Run()
}
func output(name string, args ...string) (string, error) {
	b, e := exec.Command(name, args...).Output()
	return string(b), e
}
func outputAt(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	b, e := cmd.Output()
	return string(b), e
}
func need(n string) error {
	if _, e := exec.LookPath(n); e != nil {
		return fmt.Errorf("%s not found", n)
	}
	return nil
}

func checkGitHubAccess() error {
	if _, err := exec.LookPath("gh"); err != nil {
		return errors.New("gh CLI not found. Install: https://cli.github.com")
	}
	return checkGHAuth()
}

func checkGHAuth() error {
	if err := exec.Command("gh", "auth", "status").Run(); err != nil {
		return errors.New("gh not authenticated. Run: gh auth login")
	}
	return nil
}
func show(v string) string {
	if v == "" {
		return "<unset>"
	}
	return v
}
func die(e error) {
	code := 1
	var exit exitError
	if errors.As(e, &exit) {
		code = exit.code
	}
	fmt.Fprintln(os.Stderr, "Error:", e)
	os.Exit(code)
}
func usage() {
	fmt.Printf(`start-issue v%s

Start working on a GitHub issue with git worktree and a configurable agent

Usage: start-issue <issue-url-or-number> [options]
       start-issue init [--project|--user] [--force] [options]
       start-issue setup | --setup
       start-issue update | --update
       start-issue install | --install

Arguments:
  <issue-url-or-number>      GitHub issue URL or issue number
                             Examples: 123, https://github.com/owner/repo/issues/123
  init                       Create default start-issue configuration
  setup                      Run first-run user configuration onboarding
  update                     Update the running start-issue installation
  install                    Install the latest release into ~/.local/bin

Options:
  --repo, -r <owner/repo>    Repository (default: detected from git remote)
  --base, -b <branch>        Base branch (default: main or master)
  --worktree-dir, -w <dir>   Directory for worktrees
                             Default: START_ISSUE_WORKTREE_DIR or ~/worktrees
  --flat                     Use flat worktree structure (replace / with - in path)
  --agent <name>             Agent to launch: claude, codex, kimi, pi, none
                             With init: default agent to write
  --model <name>             Model to use for the selected agent
                             With init: default model config to write
  --no-agent                 Only create worktree, do not start an agent session
  --no-claude                Compatibility alias for --no-agent
  --prompt <text>            Prompt template for the launched agent
  --prompt-file <path>       Prompt template file for the launched agent
  --improve-prompt           Ask the selected agent to improve the selected
                             prompt template and write a reviewable proposal
  --human-gate               Codex-only batch mode that resumes on HUMAN_GATE
  --human-gate-help          Show detailed help for the human-gate mode
  --prompt-output-file <path>
                             Output path for --improve-prompt proposal
  --no-init                  Skip init.sh execution
  --command, -c <cmd>        Compatibility: initial command for Claude default launch
  --ai                       Use the selected agent for branch name generation
                             Default: fast branch-name heuristics
  --project                  With init: write .start-issue config in this repo
  --user                     With init: write config in ~/.config/start-issue
  --force                    With init: overwrite existing config files
  --dry-run                  Show what would be done without executing
  --setup                    Run first-run user configuration onboarding
  --update                   Update the running start-issue installation
  --install                  Install the latest release into ~/.local/bin
  --version, -v              Show version
  --help, -h                 Show this help

Agent selection precedence:
  CLI --agent / --no-agent
  .start-issue/agent in the git root
  ~/.config/start-issue/agent
  START_ISSUE_AGENT
  built-in default: claude

Model selection precedence:
  CLI --model
  .start-issue/model in the git root
  ~/.config/start-issue/model
  START_ISSUE_MODEL
  built-in default: unset (agent CLI decides)

Prompt template precedence:
  CLI --prompt-file / --prompt
  START_ISSUE_PROMPT_FILE / START_ISSUE_PROMPT
  .start-issue/prompt.md in the git root
  ~/.config/start-issue/prompt.md
  built-in default

Prompt improvement:
  --improve-prompt uses the selected agent to generate a complete improved
  prompt template proposal. It does not overwrite the active prompt template.
  Markdown prompt files write next to the source as *.improved.md; other
  file names append .improved. Built-in and inline prompts write to
  .start-issue/prompt.improved.md by default. Use --prompt-output-file to
  choose another proposal path.

Prompt variables:
  {ISSUE_URL}, {ISSUE_NUMBER}, {ISSUE_TITLE}, {ISSUE_BODY}, {ISSUE_LABELS},
  {REPO}, {BRANCH_NAME}, {WORKTREE_PATH}, {BASE_BRANCH}

Environment variables:
  START_ISSUE_AGENT
  START_ISSUE_MODEL
  START_ISSUE_PROMPT
  START_ISSUE_PROMPT_FILE
  START_ISSUE_WORKTREE_DIR
  START_ISSUE_DUMP_PROMPT

Examples:
  start-issue 123
  start-issue https://github.com/owner/repo/issues/123
  start-issue 123 --repo owner/repo --base develop
  start-issue 123 --agent codex
  start-issue 123 --agent codex --model gpt-5.2
  start-issue 123 --agent codex --human-gate
  start-issue 123 --agent claude --model sonnet
  start-issue 123 --agent kimi --prompt-file .start-issue/prompt.md
  start-issue 123 --no-agent              # Only create worktree
  start-issue 123 --command "/debug"      # Claude command prefix
  start-issue 123 --flat                  # Flat worktree path
  start-issue 123 --dry-run
  start-issue init
  start-issue setup
  start-issue init --project --agent codex --model gpt-5.2
  start-issue init --project --agent codex
  start-issue init --user --force
  start-issue --setup
  start-issue update
  start-issue --update
  start-issue install
  start-issue --install
  start-issue --human-gate-help
`, runningVersion())
}

func printBanner() {
	fmt.Printf("start-issue v%s\n\n", runningVersion())
}

func humanGateHelp() {
	printBanner()
	fmt.Println(`Codex human-gate mode

Usage:
  start-issue <issue-url-or-number> --agent codex --human-gate
  start-issue --human-gate-help

Flow:
  The normal issue workflow creates or reuses the worktree, renders the
  prompt, and runs Codex in batch mode. The final message must contain one
  terminal status line: STATUS: DONE or STATUS: HUMAN_GATE.

Exit codes:
  0  Codex returned STATUS: DONE.
  1  Codex failed, no thread_id was captured, no recognized status was found,
     or parsing failed.
  2  Codex returned STATUS: HUMAN_GATE but automatic interactive resume failed.

State artifacts:
  <worktree>/.start-issue/runs/<timestamp>/events.jsonl
  <worktree>/.start-issue/runs/<timestamp>/last-message.txt
  <worktree>/.start-issue/runs/<timestamp>/thread-id

Recovery:
  If automatic resume fails, run:
    codex resume --include-non-interactive <thread_id>`)
}
