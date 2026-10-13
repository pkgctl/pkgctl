package tools

import (
	"reflect"
	"testing"
)

func TestParseBrewOutdated(t *testing.T) {
	output := `{
  "formulae": [
    {"name": "wget", "installed_versions": ["1.24.5"], "current_version": "1.25.0", "pinned": false, "pinned_version": null},
    {"name": "python@3.12", "installed_versions": ["3.12.6"], "current_version": "3.12.7", "pinned": true, "pinned_version": "3.12.6"}
  ],
  "casks": [
    {"name": "firefox", "installed_versions": ["130.0"], "current_version": "131.0"}
  ]
}`

	packages, err := parseBrewOutdated(output)

	if err != nil {
		t.Fatal(err)
	}
	expected := []Package{
		{Name: "wget", Current: "1.24.5", Latest: "1.25.0"},
		{Name: "firefox", Current: "130.0", Latest: "131.0"},
	}
	if !reflect.DeepEqual(packages, expected) {
		t.Errorf("Expected %v, got %v", expected, packages)
	}
}

func TestParseRustupCheck(t *testing.T) {
	output := `stable-aarch64-apple-darwin - up to date: 1.99.0 (b940084d7 2026-09-28)
nightly-aarch64-apple-darwin - update available: 1.101.0-nightly (32dba69d6 2026-10-09) -> 1.101.0-nightly (080cc2714 2026-10-11)
`

	packages := parseRustupCheck(output)

	expected := []Package{
		{Name: "nightly-aarch64-apple-darwin", Current: "1.101.0-nightly (32dba69d6 2026-10-09)", Latest: "1.101.0-nightly (080cc2714 2026-10-11)"},
	}
	if !reflect.DeepEqual(packages, expected) {
		t.Errorf("Expected %v, got %v", expected, packages)
	}
}

func TestParseMiseOutdated(t *testing.T) {
	output := `{
  "java": {"name": "java", "requested": "25.0.2", "current": "25.0.2", "bump": null, "latest": "27.0.0+35", "source": {"type": "unknown"}},
  "dotnet": {"name": "dotnet", "requested": "10.0.103", "current": "10.0.103", "bump": null, "latest": "10.0.401", "source": {"type": "unknown"}}
}`

	packages, err := parseMiseOutdated(output)

	if err != nil {
		t.Fatal(err)
	}
	expected := []Package{
		{Name: "dotnet", Current: "10.0.103", Latest: "10.0.401"},
		{Name: "java", Current: "25.0.2", Latest: "27.0.0+35"},
	}
	if !reflect.DeepEqual(packages, expected) {
		t.Errorf("Expected %v, got %v", expected, packages)
	}
}

func TestGitBlobHash(t *testing.T) {
	hash := gitBlobHash([]byte("hello\n"))

	if hash != "ce013625030ba8dba906f756967f9e9ca394464a" {
		t.Errorf("Expected ce013625030ba8dba906f756967f9e9ca394464a, got %v", hash)
	}
}

func TestUpstreamFiles(t *testing.T) {
	tree := []githubTreeEntry{
		{Path: "README.md", SHA: "readme"},
		{Path: "functions", SHA: "dir"},
		{Path: "functions/fisher.fish", SHA: "function"},
		{Path: "completions/fisher.fish", SHA: "completion"},
	}

	files := upstreamFiles(tree)

	expected := map[string]string{
		"functions/fisher.fish":   "function",
		"completions/fisher.fish": "completion",
	}
	if !reflect.DeepEqual(files, expected) {
		t.Errorf("Expected %v, got %v", expected, files)
	}
}
