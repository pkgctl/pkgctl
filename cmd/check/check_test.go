package check

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/pkgctl/pkgctl/outdated"
	"github.com/pkgctl/pkgctl/tools"
)

func TestCheckAllFinishesPastAFailure(t *testing.T) {
	broken := &tools.Tool{
		Name: "pnpm",
		ListOutdated: func(ctx context.Context) ([]tools.Package, error) {
			return nil, errors.New("no network")
		},
	}
	upToDate := &tools.Tool{
		Name: "npm",
		ListOutdated: func(ctx context.Context) ([]tools.Package, error) {
			return nil, nil
		},
	}
	outdatedTool := &tools.Tool{
		Name: "uv",
		ListOutdated: func(ctx context.Context) ([]tools.Package, error) {
			return []tools.Package{{Name: "ruff", Current: "0.5.0", Latest: "0.17.0"}}, nil
		},
	}
	report := outdated.Report{Packages: map[string][]tools.Package{}}

	err := checkAll(context.Background(), []*tools.Tool{broken, upToDate, outdatedTool}, report)

	expectedError := "checking pnpm failed"
	if err == nil || err.Error() != expectedError {
		t.Errorf("Expected %q, got %v", expectedError, err)
	}
	expected := map[string][]tools.Package{"uv": {{Name: "ruff", Current: "0.5.0", Latest: "0.17.0"}}}
	if !reflect.DeepEqual(report.Packages, expected) {
		t.Errorf("Expected %v, got %v", expected, report.Packages)
	}
}
