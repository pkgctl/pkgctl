package outdated

import (
	"reflect"
	"testing"

	"github.com/pkgctl/pkgctl/tools"
)

func TestSummary(t *testing.T) {
	report := Report{Packages: map[string][]tools.Package{
		"rustup":   {{Name: "nightly-aarch64-apple-darwin"}},
		"Homebrew": {{Name: "wget"}, {Name: "firefox"}},
	}}

	summary := report.Summary()

	expected := "3 outdated packages (Homebrew 2, rustup 1). Run `pkgctl update`."
	if summary != expected {
		t.Errorf("Expected %q, got %q", expected, summary)
	}
}

func TestDescribeWithoutVersions(t *testing.T) {
	packages := []tools.Package{{Name: "pure-fish/pure"}}

	description := Describe(tools.Fisher, packages)

	expected := "  pure-fish/pure\n↑ 1 outdated. Run `fisher update`.\n"
	if description != expected {
		t.Errorf("Expected %q, got %q", expected, description)
	}
}

func TestClearPackage(t *testing.T) {
	report := Report{Packages: map[string][]tools.Package{
		"Go": {{Name: "golang.org/x/tools/gopls"}, {Name: "honnef.co/go/tools/cmd/staticcheck"}},
	}}

	report.Clear("Go", "golang.org/x/tools/gopls")

	expected := []tools.Package{{Name: "honnef.co/go/tools/cmd/staticcheck"}}
	if !reflect.DeepEqual(report.Packages["Go"], expected) {
		t.Errorf("Expected %v, got %v", expected, report.Packages["Go"])
	}
}

func TestClearPackageBeforeAnyCheck(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	report, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	report.Clear("Go", "golang.org/x/tools/gopls")

	if len(report.Packages["Go"]) != 0 {
		t.Errorf("Expected no Go packages, got %v", report.Packages["Go"])
	}
}

func TestDescribeWithDisplayName(t *testing.T) {
	packages := []tools.Package{{Name: "io.tailscale.ipn.macos", DisplayName: "Tailscale", Current: "1.102.4", Latest: "1.104.1"}}

	description := Describe(tools.Mas, packages)

	expected := "  Tailscale (io.tailscale.ipn.macos) 1.102.4 → 1.104.1\n↑ 1 outdated. Run `mas update`.\n"
	if description != expected {
		t.Errorf("Expected %q, got %q", expected, description)
	}
}
