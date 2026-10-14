package tools

import (
	"context"
	"reflect"
	"testing"
)

func TestUpdateNamesOutdatedPackages(t *testing.T) {
	t.Setenv("SHELL", "")
	tool := &Tool{
		Command:           "pnpm",
		UpdatePackageArgs: []string{"update", "--global"},
		ListOutdated: func(ctx context.Context) ([]Package, error) {
			return []Package{{Name: "cowsay"}, {Name: "typescript"}}, nil
		},
	}

	cmds, err := tool.Update(context.Background(), "")

	if err != nil {
		t.Fatal(err)
	}
	if len(cmds) != 1 {
		t.Fatalf("Expected one command, got %v", len(cmds))
	}
	expected := []string{"pnpm", "update", "--global", "cowsay", "typescript"}
	if !reflect.DeepEqual(cmds[0].Args, expected) {
		t.Errorf("Expected %v, got %v", expected, cmds[0].Args)
	}
}

func TestUpdateWithNothingOutdated(t *testing.T) {
	tool := &Tool{
		Command:           "pnpm",
		UpdatePackageArgs: []string{"update", "--global"},
		ListOutdated: func(ctx context.Context) ([]Package, error) {
			return nil, nil
		},
	}

	cmds, err := tool.Update(context.Background(), "")

	if err != nil {
		t.Fatal(err)
	}
	if len(cmds) != 0 {
		t.Errorf("Expected no commands, got %v", cmds)
	}
}
