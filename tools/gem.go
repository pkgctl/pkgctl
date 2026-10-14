package tools

import (
	"context"
	"regexp"
	"slices"
)

var Gem = &Tool{
	Name:              "RubyGems",
	Command:           "gem",
	Description:       "The Ruby package manager",
	VersionArgs:       []string{"--version"},
	UpdatePackageArgs: []string{"update"},
	ListOutdated:      listGemOutdated,
}

// rubyBundledGems are the gems Ruby installs alongside itself, every gem listed in gems/bundled_gems
// in Ruby's source from 3.3 through its development branch, which covers the earlier versions' lists too.
var rubyBundledGems = []string{
	"abbrev", "base64", "benchmark", "bigdecimal", "csv", "debug", "drb", "fiddle", "getoptlong", "irb",
	"logger", "matrix", "minitest", "mutex_m", "net-ftp", "net-imap", "net-pop", "net-smtp", "nkf", "observer",
	"ostruct", "power_assert", "prime", "prism", "pstore", "racc", "rake", "rbs", "rdoc", "readline",
	"reline", "repl_type_completor", "resolv-replace", "rexml", "rinda", "rss", "syslog", "test-unit", "tsort",
	"typeprof", "win32-registry", "win32ole",
}

// listGemOutdated lists the outdated gems that don't ship with Ruby, as a default or bundled gem,
// since those come with Ruby itself and its installer, such as Homebrew or mise, updates them.
func listGemOutdated(ctx context.Context) ([]Package, error) {
	listOutput, err := ExecContext(ctx, "gem", "list").Output()
	if err != nil {
		return nil, err
	}
	outdatedOutput, err := ExecContext(ctx, "gem", "outdated").Output()
	if err != nil {
		return nil, err
	}

	defaultGems := parseGemDefaults(string(listOutput))
	var packages []Package
	for _, outdatedPackage := range parseGemOutdated(string(outdatedOutput)) {
		if slices.Contains(defaultGems, outdatedPackage.Name) || slices.Contains(rubyBundledGems, outdatedPackage.Name) {
			continue
		}
		packages = append(packages, outdatedPackage)
	}
	return packages, nil
}

// A `gem outdated` line, such as "bigdecimal (4.0.1 < 4.1.3)".
var gemOutdatedLine = regexp.MustCompile(`(?m)^(\S+) \((\S+) < (\S+)\)$`)

func parseGemOutdated(output string) []Package {
	var packages []Package
	for _, match := range gemOutdatedLine.FindAllStringSubmatch(output, -1) {
		packages = append(packages, Package{Name: match[1], Current: match[2], Latest: match[3]})
	}
	return packages
}

// A `gem list` line for a gem that ships with Ruby as a default gem, such as "date (3.6.0, default: 3.5.1)".
var gemDefaultLine = regexp.MustCompile(`(?m)^(\S+) \(.*\bdefault: `)

// parseGemDefaults parses `gem list` into the names of Ruby's default gems.
func parseGemDefaults(output string) []string {
	var names []string
	for _, match := range gemDefaultLine.FindAllStringSubmatch(output, -1) {
		names = append(names, match[1])
	}
	return names
}
