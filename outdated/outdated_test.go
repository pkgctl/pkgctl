package outdated

import (
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

func TestDescribeWithDisplayName(t *testing.T) {
	packages := []tools.Package{{Name: "io.tailscale.ipn.macos", DisplayName: "Tailscale", Current: "1.102.4", Latest: "1.104.1"}}

	description := Describe(tools.Mas, packages)

	expected := "  Tailscale (io.tailscale.ipn.macos) 1.102.4 → 1.104.1\n↑ 1 outdated. Run `mas update`.\n"
	if description != expected {
		t.Errorf("Expected %q, got %q", expected, description)
	}
}
