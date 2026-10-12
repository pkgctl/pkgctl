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

	expected := "\033[1;33m3 outdated packages\033[0m (Homebrew 2, rustup 1). Run `pkgctl update`."
	if summary != expected {
		t.Errorf("Expected %q, got %q", expected, summary)
	}
}
