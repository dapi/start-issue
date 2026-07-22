package main

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestParseIssue(t *testing.T) {
	number, repo, err := parseIssue("https://github.com/dapi/start-issue/issues/34", "")
	if err != nil || number != "34" || repo != "dapi/start-issue" {
		t.Fatalf("got %q %q %v", number, repo, err)
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
		{"", "", "feature/issue-34-work"},
	}
	for _, test := range tests {
		if got := branchName("34", test.title, test.labels); got != test.want {
			t.Errorf("branchName(%q, %q) = %q, want %q", test.title, test.labels, got, test.want)
		}
	}
}

func TestRender(t *testing.T) {
	if got := render("{REPO} #{ISSUE_NUMBER}", map[string]string{"REPO": "dapi/start-issue", "ISSUE_NUMBER": "34"}); got != "dapi/start-issue #34" {
		t.Fatalf("got %q", got)
	}
}

func TestReleaseAssetName(t *testing.T) {
	if got := releaseAssetName("darwin", "arm64"); got != "start-issue-darwin-arm64" {
		t.Fatalf("got %q", got)
	}
	if got := releaseAssetName("windows", "amd64"); got != "start-issue-windows-amd64.exe" {
		t.Fatalf("got %q", got)
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
	if compareVersions("v1.2.0", "1.2.0") != 0 || compareVersions("1.2.1", "1.2.0") <= 0 || compareVersions("1.1.9", "v1.2.0") >= 0 {
		t.Fatal("unexpected version ordering")
	}
}
