package main

import (
	"bufio"
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
	"strconv"
	"strings"
	"time"
)

var version = "2.0.0"

type options struct {
	repo, base, worktreeDir, agent, model, promptFile, prompt, command       string
	promptOutput                                                             string
	issue                                                                    string
	dryRun, noInit, flat, ai, improvePrompt, humanGate, project, user, force bool
	mode                                                                     string
}

type issue struct {
	Title, Body string
	Labels      []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

func main() {
	o, err := parse(os.Args[1:])
	if err != nil {
		die(err)
	}
	if o.mode != "" {
		if err := runMode(o); err != nil {
			die(err)
		}
		return
	}
	if o.issue == "" {
		usage()
		return
	}
	if err := run(o); err != nil {
		die(err)
	}
}

func parse(args []string) (options, error) {
	o := options{worktreeDir: os.Getenv("START_ISSUE_WORKTREE_DIR")}
	var err error
	if o.worktreeDir == "" {
		home, _ := os.UserHomeDir()
		o.worktreeDir = filepath.Join(home, "worktrees")
	}
	for len(args) > 0 {
		a := args[0]
		args = args[1:]
		value := func() (string, error) {
			if len(args) == 0 || strings.HasPrefix(args[0], "-") {
				return "", fmt.Errorf("%s requires a value.", a)
			}
			v := args[0]
			args = args[1:]
			return v, nil
		}
		switch a {
		case "--help", "-h":
			usage()
			os.Exit(0)
		case "--version", "-v":
			fmt.Printf("start-issue v%s\n", version)
			os.Exit(0)
		case "--repo", "-r":
			o.repo, err = value()
		case "--base", "-b":
			o.base, err = value()
		case "--worktree-dir", "-w":
			o.worktreeDir, err = value()
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
		case "init":
			o.mode = "init"
		case "setup", "--setup":
			o.mode = "setup"
		case "update", "--update":
			o.mode = "update"
		case "install", "--install":
			o.mode = "install"
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
	return o, nil
}

func run(o options) error {
	root, err := output("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return errors.New("Not in a git repository")
	}
	root = strings.TrimSpace(root)
	if err := maybeRunFirstRunOnboarding(); err != nil {
		return err
	}
	agent, agentSource, err := resolveAgent(root, o.agent)
	if err != nil {
		return err
	}
	if o.humanGate && agent != "codex" {
		return fmt.Errorf("--human-gate requires agent 'codex'. Current agent: %s.", agent)
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
	prompt, promptSource, err := resolvePrompt(root, agent, o)
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
	if err := need("gh"); err != nil {
		return err
	}
	if err := need("git"); err != nil {
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
	if _, err := exec.LookPath("zellij-tab-status"); err == nil {
		if o.dryRun {
			fmt.Printf("   [DRY-RUN] Would run: zellij-tab-status --set-name #%s\n", number)
		} else {
			_ = command("zellij-tab-status", "--set-name", "#"+number)
		}
	}
	if o.improvePrompt {
		return improvePrompt(root, agent, model, prompt, promptSource, o, in, repo, number, strings.Join(labels, ", "))
	}
	branch := branchName(number, in.Title, strings.Join(labels, ", "))
	if o.ai {
		if generated, err := aiBranchName(agent, model, root, number, in.Title, strings.Join(labels, ", ")); err == nil && regexp.MustCompile(`^(feature|fix|hotfix|refactor|docs|test|chore)/issue-[0-9]+-[a-z0-9][a-z0-9-]*$`).MatchString(generated) {
			branch = generated
			fmt.Printf("   Branch: %s (ai:%s)\n", branch, agent)
		} else {
			fmt.Printf("   Could not generate branch name with %s; using fast fallback\n", agent)
		}
	}
	name := branch
	if o.flat {
		name = strings.ReplaceAll(name, "/", "-")
	}
	worktree := filepath.Join(o.worktreeDir, name)
	issueURL := fmt.Sprintf("https://github.com/%s/issues/%s", repo, number)
	rendered := render(prompt, map[string]string{"ISSUE_URL": issueURL, "ISSUE_NUMBER": number, "ISSUE_TITLE": in.Title, "ISSUE_BODY": in.Body, "ISSUE_LABELS": strings.Join(labels, ", "), "REPO": repo, "BRANCH_NAME": branch, "WORKTREE_PATH": worktree, "BASE_BRANCH": o.base})
	fmt.Printf("Agent: %s\nAgent source: %s\nModel: %s\nModel source: %s\nWorktree directory: %s\nPrompt source: %s\n\n", agent, agentSource, show(model), modelSource, o.worktreeDir, promptSource)
	fmt.Printf("🔍 Fetching issue #%s from %s...\n   Title: %s\n", number, repo, in.Title)
	fmt.Printf("   Branch: %s (fast)\n📁 Creating worktree...\n   Path: %s\n   Base: %s\n", branch, worktree, o.base)
	if o.dryRun {
		fmt.Printf("   [DRY-RUN] Would run: git worktree add -b %s %s %s\n", branch, worktree, o.base)
		printLaunch(agent, model, worktree, rendered)
		return nil
	}
	if _, err := os.Stat(worktree); err == nil {
		if worktreeBranch(worktree) != "refs/heads/"+branch {
			return fmt.Errorf("Cannot reuse worktree path '%s': it does not belong to branch '%s'.", worktree, branch)
		}
		if !o.noInit {
			runInit(worktree)
		}
		return launchSelected(o, agent, model, worktree, rendered)
	}
	if branchWorktree(branch) != "" {
		return fmt.Errorf("Branch '%s' already exists in worktree %s.", branch, branchWorktree(branch))
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
		runInit(worktree)
	}
	return launchSelected(o, agent, model, worktree, rendered)
}

func runInit(worktree string) {
	if init := filepath.Join(worktree, "init.sh"); fileExists(init) {
		_ = commandAt(worktree, "bash", "./init.sh")
	}
}
func fileExists(path string) bool { _, err := os.Stat(path); return err == nil }
func maybeRunFirstRunOnboarding() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".config", "start-issue")
	if _, err := os.Stat(dir); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("No start-issue user configuration found. Run setup now? [Y/n] ")
	answer, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(answer)) == "n" {
		return os.MkdirAll(dir, 0755)
	}
	return setupMode(home)
}

func runMode(o options) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	root, _ := output("git", "rev-parse", "--show-toplevel")
	root = strings.TrimSpace(root)
	if o.mode == "update" {
		return updateMode(o)
	}
	if o.mode == "install" {
		return installMode()
	}
	if o.mode == "setup" {
		return setupMode(home)
	}
	dir := filepath.Join(home, ".config", "start-issue")
	if o.mode == "init" && o.project {
		if root == "" {
			return errors.New("--project requires a git repository")
		}
		dir = filepath.Join(root, ".start-issue")
	} else if o.mode == "init" && !o.user && root != "" {
		dir = filepath.Join(root, ".start-issue")
	}
	if o.dryRun {
		fmt.Printf("[DRY-RUN] Would create configuration in: %s\n", dir)
		return nil
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	agent := o.agent
	if agent == "" {
		agent = "claude"
	}
	if err := writeConfig(filepath.Join(dir, "agent"), agent+"\n", o.force); err != nil {
		return err
	}
	if o.model != "" {
		if err := writeConfig(filepath.Join(dir, "model"), o.model+"\n", o.force); err != nil {
			return err
		}
	}
	prompt := o.prompt
	if o.promptFile != "" {
		b, err := os.ReadFile(o.promptFile)
		if err != nil {
			return err
		}
		prompt = string(b)
	}
	if prompt == "" {
		if agent == "claude" {
			prompt = "/task-router:route-task {ISSUE_URL}"
		} else {
			prompt = "Implement GitHub issue {ISSUE_URL} in this worktree."
		}
	}
	if err := writeConfig(filepath.Join(dir, "prompt.md"), prompt+"\n", o.force); err != nil {
		return err
	}
	fmt.Printf("Wrote start-issue configuration: %s\n", dir)
	return nil
}

func setupMode(home string) error {
	dir := filepath.Join(home, ".config", "start-issue")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Select default agent: 1) claude  2) codex  3) kimi  4) pi  5) skip")
	fmt.Print("Choice [1]: ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	agents := map[string]string{"": "claude", "1": "claude", "2": "codex", "3": "kimi", "4": "pi"}
	if agent, ok := agents[choice]; ok {
		if err := writeConfig(filepath.Join(dir, "agent"), agent+"\n", false); err != nil {
			return err
		}
	} else if choice != "5" {
		return errors.New("invalid setup choice")
	}
	fmt.Print("Save a default prompt? [Y/n] ")
	answer, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(answer)) != "n" {
		agent := agents[choice]
		prompt := "Implement GitHub issue {ISSUE_URL} in this worktree."
		if agent == "claude" {
			prompt = "/task-router:route-task {ISSUE_URL}"
		}
		if err := writeConfig(filepath.Join(dir, "prompt.md"), prompt+"\n", false); err != nil {
			return err
		}
	}
	fmt.Printf("Wrote start-issue configuration: %s\n", dir)
	return nil
}

func installMode() error {
	if runtime.GOOS == "windows" {
		return errors.New("Windows installation is manual: download start-issue-windows-amd64.exe from the latest release")
	}
	data, err := output("gh", "api", "repos/dapi/start-issue/releases/latest")
	if err != nil {
		return errors.New("Could not fetch the latest start-issue release")
	}
	var release githubRelease
	if err := json.Unmarshal([]byte(data), &release); err != nil {
		return err
	}
	name := releaseAssetName(runtime.GOOS, runtime.GOARCH)
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
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	target := filepath.Join(home, ".local", "bin", "start-issue")
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(target, binary, 0755); err != nil {
		return err
	}
	fmt.Printf("Installed start-issue v%s at: %s\n", strings.TrimPrefix(release.TagName, "v"), target)
	return nil
}

func writeConfig(path, content string, force bool) error {
	if _, err := os.Stat(path); err == nil && !force {
		return nil
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func updateMode(o options) error {
	if runtime.GOOS == "windows" {
		fmt.Println("Windows update is manual: download start-issue-windows-amd64.exe from the latest release and replace the executable on PATH.")
		return nil
	}
	data, err := output("gh", "api", "repos/dapi/start-issue/releases/latest")
	if err != nil {
		return errors.New("Could not fetch the latest start-issue release")
	}
	var release githubRelease
	if err := json.Unmarshal([]byte(data), &release); err != nil {
		return fmt.Errorf("decode latest release: %w", err)
	}
	if compareVersions(version, release.TagName) >= 0 {
		fmt.Printf("start-issue is already up to date (%s).\n", version)
		return nil
	}
	assetName := releaseAssetName(runtime.GOOS, runtime.GOARCH)
	assetURL, checksumURL := release.assetURLs(assetName)
	if assetURL == "" || checksumURL == "" {
		return fmt.Errorf("latest release does not contain %s and checksums.txt", assetName)
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
	target, err := os.Executable()
	if err != nil {
		return err
	}
	temporary := target + ".new"
	if err := os.WriteFile(temporary, binary, 0755); err != nil {
		return err
	}
	if err := os.Rename(temporary, target); err != nil {
		_ = os.Remove(temporary)
		return err
	}
	fmt.Printf("Updated start-issue at: %s\nVersion: start-issue v%s\n", target, strings.TrimPrefix(release.TagName, "v"))
	return nil
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
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

func releaseAssetName(goos, goarch string) string {
	name := fmt.Sprintf("start-issue-%s-%s", goos, goarch)
	if goos == "windows" {
		return name + ".exe"
	}
	return name
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

func compareVersions(left, right string) int {
	parse := func(value string) [3]int {
		var result [3]int
		for index, part := range strings.Split(strings.TrimPrefix(value, "v"), ".") {
			if index == len(result) {
				break
			}
			result[index], _ = strconv.Atoi(regexp.MustCompile(`^[0-9]+`).FindString(part))
		}
		return result
	}
	a, b := parse(left), parse(right)
	for i := range a {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

func resolveAgent(root, cli string) (string, string, error) {
	home, _ := os.UserHomeDir()
	v, s, e := resolve(cli, filepath.Join(root, ".start-issue", "agent"), filepath.Join(home, ".config", "start-issue", "agent"), "START_ISSUE_AGENT", "claude")
	if e != nil {
		return "", "", e
	}
	switch v {
	case "claude", "codex", "kimi", "pi", "none":
		return v, s, nil
	}
	return "", "", fmt.Errorf("Unknown agent: %s. Valid agents: claude, codex, kimi, pi, none.", v)
}
func resolveModel(root, cli string) (string, string, error) {
	home, _ := os.UserHomeDir()
	v, s, e := resolve(cli, filepath.Join(root, ".start-issue", "model"), filepath.Join(home, ".config", "start-issue", "model"), "START_ISSUE_MODEL", "")
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
		if b, e := os.ReadFile(p); e == nil {
			return first(string(b)), p, nil
		} else if !os.IsNotExist(e) {
			return "", "", e
		}
	}
	if v := strings.TrimSpace(os.Getenv(env)); v != "" {
		return v, env, nil
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
func resolvePrompt(root, agent string, o options) (string, string, error) {
	if o.prompt != "" {
		return o.prompt, "CLI --prompt", nil
	}
	if o.promptFile != "" {
		b, e := os.ReadFile(o.promptFile)
		return string(b), "CLI --prompt-file: " + o.promptFile, e
	}
	if os.Getenv("START_ISSUE_PROMPT_FILE") != "" && os.Getenv("START_ISSUE_PROMPT") != "" {
		return "", "", errors.New("Use either START_ISSUE_PROMPT_FILE or START_ISSUE_PROMPT, not both.")
	}
	if path := os.Getenv("START_ISSUE_PROMPT_FILE"); path != "" {
		b, e := os.ReadFile(path)
		return string(b), "START_ISSUE_PROMPT_FILE: " + path, e
	}
	if value := os.Getenv("START_ISSUE_PROMPT"); value != "" {
		return value, "START_ISSUE_PROMPT", nil
	}
	home, _ := os.UserHomeDir()
	for _, path := range []string{filepath.Join(root, ".start-issue", "prompt.md"), filepath.Join(home, ".config", "start-issue", "prompt.md")} {
		if b, e := os.ReadFile(path); e == nil {
			return string(b), path, nil
		} else if !os.IsNotExist(e) {
			return "", "", e
		}
	}
	if agent == "claude" {
		if o.command != "" {
			return o.command + " {ISSUE_URL}", "built-in Claude command", nil
		}
		return "/task-router:route-task {ISSUE_URL}", "built-in Claude command", nil
	}
	return "Implement GitHub issue {ISSUE_URL} in this worktree.", "built-in portable prompt", nil
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
	v = strings.TrimSpace(strings.TrimSuffix(v, ".git"))
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
	l := strings.ToLower(labels)
	switch {
	case containsAny(l, "hotfix", "critical", "urgent"):
		kind = "hotfix"
	case containsAny(l, "bug", "fix", "bugfix", "error"):
		kind = "fix"
	case containsAny(l, "docs", "documentation"):
		kind = "docs"
	case containsAny(l, "refactor", "tech-debt", "cleanup", "technical"):
		kind = "refactor"
	case containsAny(l, "test", "testing", "tests"):
		kind = "test"
	case containsAny(l, "chore", "ci", "build", "infra"):
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
	prompt := fmt.Sprintf("Git branch name for issue #%s: %q (labels: %s). Reply with ONLY {type}/issue-%s-{kebab-case-name}.", number, title, labels, number)
	args := helperArgs(agent, model, root, prompt)
	result, err := output(args[0], args[1:]...)
	if err != nil {
		return "", err
	}
	lines := strings.Fields(strings.Trim(strings.TrimSpace(result), "`\""))
	if len(lines) == 0 {
		return "", errors.New("empty branch")
	}
	return lines[len(lines)-1], nil
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
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "work"
	}
	if len(slug) > 40 {
		slug = strings.Trim(slug[:40], "-")
	}
	return slug
}

var cyrillicTransliteration = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "yo", 'ж': "zh", 'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m", 'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u", 'ф': "f", 'х': "kh", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "shch", 'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
}

func render(s string, m map[string]string) string {
	for k, v := range m {
		s = strings.ReplaceAll(s, "{"+k+"}", v)
	}
	return s
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
	value, err := output("git", "worktree", "list", "--porcelain")
	if err != nil {
		return ""
	}
	current := ""
	for _, line := range strings.Split(value, "\n") {
		if strings.HasPrefix(line, "worktree ") {
			current = strings.TrimPrefix(line, "worktree ")
		}
		if current == path && strings.HasPrefix(line, "branch ") {
			return strings.TrimPrefix(line, "branch ")
		}
	}
	return ""
}
func launchSelected(o options, agent, model, worktree, prompt string) error {
	if o.humanGate {
		return humanGate(model, worktree, prompt, o.dryRun)
	}
	return launch(agent, model, worktree, prompt)
}
func improvePrompt(root, agent, model, prompt, source string, o options, in issue, repo, number, labels string) error {
	if agent == "none" {
		return errors.New("--improve-prompt requires an agent. Use --agent claude, codex, kimi, or pi.")
	}
	outputPath := o.promptOutput
	if outputPath == "" {
		if o.promptFile != "" {
			outputPath = strings.TrimSuffix(o.promptFile, filepath.Ext(o.promptFile)) + ".improved.md"
		} else {
			outputPath = filepath.Join(root, ".start-issue", "prompt.improved.md")
		}
	}
	if o.dryRun {
		fmt.Printf("📝 Improving prompt template...\n   Prompt source: %s\n   Proposal path: %s\n   [DRY-RUN] Would ask %s to generate an improved prompt proposal.\n", source, outputPath, agent)
		return nil
	}
	if _, err := os.Stat(outputPath); err == nil {
		return fmt.Errorf("Prompt improvement output already exists: %s", outputPath)
	}
	request := fmt.Sprintf("Improve the following start-issue prompt template. Return ONLY the complete improved prompt template.\n\nRepository: %s\nIssue #%s: %s\nLabels: %s\nBody:\n%s\n\nPrompt:\n%s", repo, number, in.Title, labels, in.Body, prompt)
	args := helperArgs(agent, model, root, request)
	result, err := output(args[0], args[1:]...)
	if err != nil {
		return fmt.Errorf("Could not generate improved prompt with %s", agent)
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, []byte(strings.TrimSpace(result)+"\n"), 0644); err != nil {
		return err
	}
	fmt.Printf("📝 Prompt improvement written: %s\n", outputPath)
	return nil
}
func humanGate(model, worktree, prompt string, dryRun bool) error {
	runID := os.Getenv("START_ISSUE_RUN_ID")
	if runID == "" {
		runID = time.Now().Format("20060102-150405")
	}
	dir := filepath.Join(worktree, ".start-issue", "runs", runID)
	events, last := filepath.Join(dir, "events.jsonl"), filepath.Join(dir, "last-message.txt")
	args := []string{"exec", "--cd", worktree, "--ask-for-approval", "never", "--sandbox", "workspace-write", "--json", "--output-last-message", last, "-"}
	if model != "" {
		args = append([]string{"exec", "--model", model}, args[1:]...)
	}
	if dryRun {
		fmt.Printf("   [DRY-RUN] Would run: codex %s\n", strings.Join(args, " "))
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
	cmd.Dir = worktree
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Stdout = file
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
	body, err := os.ReadFile(last)
	if err != nil {
		return fmt.Errorf("No recognized final status found. Inspect: %s", last)
	}
	status := ""
	for _, line := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(line, "STATUS:") {
			status = strings.TrimSpace(strings.TrimPrefix(line, "STATUS:"))
			break
		}
	}
	if status == "DONE" {
		fmt.Println("✅ Codex finished with STATUS: DONE")
		return nil
	}
	if status == "HUMAN_GATE" {
		eventsBody, err := os.ReadFile(events)
		if err != nil {
			return err
		}
		threadID := ""
		for _, line := range strings.Split(string(eventsBody), "\n") {
			var event struct {
				Type     string `json:"type"`
				ThreadID string `json:"thread_id"`
			}
			if json.Unmarshal([]byte(line), &event) == nil && event.Type == "thread.started" {
				threadID = event.ThreadID
				break
			}
		}
		if threadID == "" {
			return fmt.Errorf("Codex human-gate run did not capture thread_id. Inspect: %s", events)
		}
		if err := os.WriteFile(filepath.Join(dir, "thread-id"), []byte(threadID+"\n"), 0644); err != nil {
			return err
		}
		return command("codex", "resume", "--include-non-interactive", threadID)
	}
	return fmt.Errorf("No recognized final status found. Inspect: %s", last)
}
func printLaunch(a, m, w, p string) {
	if a == "none" {
		fmt.Printf("   Agent: none\n   Model: %s\n   [DRY-RUN] Would prepare worktree without launching an agent\n", show(m))
		return
	}
	fmt.Printf("   Agent: %s\n   Model: %s\n   [DRY-RUN] Would run: %s\n", a, show(m), strings.Join(launchArgs(a, m, w, p), " "))
}
func launch(a, m, w, p string) error {
	if a == "none" {
		fmt.Printf("✅ Worktree ready at: %s\n", w)
		return nil
	}
	return commandAt(w, launchArgs(a, m, w, p)...)
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
		return append(x, "--work-dir", w, "--yolo", "-p", p)
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
		return append(args, "--work-dir", root, "--quiet", "-p", prompt)
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
func need(n string) error {
	if _, e := exec.LookPath(n); e != nil {
		return fmt.Errorf("%s not found", n)
	}
	return nil
}
func show(v string) string {
	if v == "" {
		return "<unset>"
	}
	return v
}
func die(e error) { fmt.Fprintln(os.Stderr, "Error:", e); os.Exit(1) }
func usage() {
	fmt.Printf("start-issue v%s\n\nUsage: start-issue <issue-url-or-number> [options]\n", version)
}

func humanGateHelp() {
	fmt.Println("Codex human-gate mode\n\nUsage: start-issue <issue> --agent codex --human-gate\n\nThe final Codex message must start with STATUS: DONE or STATUS: HUMAN_GATE.")
}
