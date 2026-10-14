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

func TestParseNodeOutdated(t *testing.T) {
	output := `{
  "npm": {"current": "11.19.0", "wanted": "12.2.0", "latest": "12.2.0", "dependent": "global"},
  "cowsay": {"current": "1.4.0", "latest": "1.6.0", "wanted": "1.4.0", "isDeprecated": false}
}`

	packages, err := parseNodeOutdated(output)

	if err != nil {
		t.Fatal(err)
	}
	expected := []Package{
		{Name: "cowsay", Current: "1.4.0", Latest: "1.6.0"},
		{Name: "npm", Current: "11.19.0", Latest: "12.2.0"},
	}
	if !reflect.DeepEqual(packages, expected) {
		t.Errorf("Expected %v, got %v", expected, packages)
	}
}

func TestParseGemOutdated(t *testing.T) {
	output := `bigdecimal (4.0.1 < 4.1.3)
cocoapods (1.16.2 < 1.17.0)
`

	packages := parseGemOutdated(output)

	expected := []Package{
		{Name: "bigdecimal", Current: "4.0.1", Latest: "4.1.3"},
		{Name: "cocoapods", Current: "1.16.2", Latest: "1.17.0"},
	}
	if !reflect.DeepEqual(packages, expected) {
		t.Errorf("Expected %v, got %v", expected, packages)
	}
}

func TestParseGemDefaults(t *testing.T) {
	output := `bundler (4.0.22, default: 4.0.20)
cocoapods (1.16.2)
did_you_mean (default: 2.0.0)
rake (13.4.2, 13.3.1)
`

	names := parseGemDefaults(output)

	expected := []string{"bundler", "did_you_mean"}
	if !reflect.DeepEqual(names, expected) {
		t.Errorf("Expected %v, got %v", expected, names)
	}
}

func TestParsePipOutdated(t *testing.T) {
	output := `[{"name": "requests", "version": "2.31.0", "latest_version": "2.34.2", "latest_filetype": "wheel"}]`

	packages, err := parsePipOutdated(output)

	if err != nil {
		t.Fatal(err)
	}
	expected := []Package{
		{Name: "requests", Current: "2.31.0", Latest: "2.34.2"},
	}
	if !reflect.DeepEqual(packages, expected) {
		t.Errorf("Expected %v, got %v", expected, packages)
	}
}

func TestParsePipxOutdated(t *testing.T) {
	output := `{
  "command": ["list"],
  "data": {
    "packages": [
      {"environment": "cowsay", "injected": false, "latest_version": "6.1", "package": "cowsay", "pinned": false, "version": "5.0"},
      {"environment": "black", "injected": false, "latest_version": "26.9.0", "package": "black", "pinned": true, "version": "24.1.0"}
    ],
    "packages_checked": 2,
    "skipped": []
  },
  "errors": [],
  "exit_code": 0,
  "pipx_result_version": "1",
  "status": "success"
}`

	packages, err := parsePipxOutdated(output)

	if err != nil {
		t.Fatal(err)
	}
	expected := []Package{
		{Name: "cowsay", Current: "5.0", Latest: "6.1"},
	}
	if !reflect.DeepEqual(packages, expected) {
		t.Errorf("Expected %v, got %v", expected, packages)
	}
}

func TestParseUvOutdated(t *testing.T) {
	output := `ruff v0.5.0 [latest: 0.17.0]
- ruff
`

	packages := parseUvOutdated(output)

	expected := []Package{
		{Name: "ruff", Current: "0.5.0", Latest: "0.17.0"},
	}
	if !reflect.DeepEqual(packages, expected) {
		t.Errorf("Expected %v, got %v", expected, packages)
	}
}

func TestParseDotnetTools(t *testing.T) {
	output := `{"version":1,"data":[{"packageId":"docfx","version":"2.78.5","commands":["docfx"]},{"packageId":"microsoft.dotnet.apicompat.tool","version":"10.0.400","commands":["apicompat"]}]}`

	tools, err := parseDotnetTools(output)

	if err != nil {
		t.Fatal(err)
	}
	expected := []Package{
		{Name: "docfx", Current: "2.78.5"},
		{Name: "microsoft.dotnet.apicompat.tool", Current: "10.0.400"},
	}
	if !reflect.DeepEqual(tools, expected) {
		t.Errorf("Expected %v, got %v", expected, tools)
	}
}

func TestParseNugetLatest(t *testing.T) {
	output := `{
  "version": 2,
  "problems": [],
  "searchResult": [
    {
      "sourceName": "nuget.org",
      "packages": [
        {"id": "docfx", "version": "2.80.1"},
        {"id": "docfx", "version": "2.81.0"}
      ]
    }
  ]
}`

	latest, err := parseNugetLatest(output)

	if err != nil {
		t.Fatal(err)
	}
	if latest != "2.81.0" {
		t.Errorf("Expected 2.81.0, got %v", latest)
	}
}

func TestParseGoModules(t *testing.T) {
	output := `{
	"Path": "github.com/go-delve/delve",
	"Version": "v1.27.2",
	"Time": "2026-09-30T14:02:11Z"
}
{
	"Path": "github.com/rhysd/actionlint",
	"Version": "v1.7.12",
	"Time": "2026-03-02T09:41:55Z"
}
`

	versions, err := parseGoModules(output)

	if err != nil {
		t.Fatal(err)
	}
	expected := map[string]string{
		"github.com/go-delve/delve":   "v1.27.2",
		"github.com/rhysd/actionlint": "v1.7.12",
	}
	if !reflect.DeepEqual(versions, expected) {
		t.Errorf("Expected %v, got %v", expected, versions)
	}
}

func TestParseMasOutdated(t *testing.T) {
	output := `{
  "adamID": 497799835,
  "bundleID": "com.apple.dt.Xcode",
  "name": "Xcode",
  "newVersion": "27.1",
  "version": "27.0.1"
}
{
  "adamID": 409183694,
  "bundleID": "com.apple.iWork.Keynote",
  "name": "Keynote",
  "newVersion": "15.2",
  "version": "15.1"
}`

	packages, err := parseMasOutdated(output)

	if err != nil {
		t.Fatal(err)
	}
	expected := []Package{
		{Name: "com.apple.dt.Xcode", DisplayName: "Xcode", Current: "27.0.1", Latest: "27.1"},
		{Name: "com.apple.iWork.Keynote", DisplayName: "Keynote", Current: "15.1", Latest: "15.2"},
	}
	if !reflect.DeepEqual(packages, expected) {
		t.Errorf("Expected %v, got %v", expected, packages)
	}
}
