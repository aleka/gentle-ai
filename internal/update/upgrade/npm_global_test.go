package upgrade

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/system"
	"github.com/gentleman-programming/gentle-ai/v2/internal/update"
)

// TestRunStrategyNpmGlobalUpgrade verifies the npm-global strategy runs exactly
// `npm install -g <pkg>@latest` through the execCommand seam.
func TestRunStrategyNpmGlobalUpgrade(t *testing.T) {
	origExecCommand := execCommand
	t.Cleanup(func() { execCommand = origExecCommand })

	var gotName string
	var gotArgs []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		gotName = name
		gotArgs = args
		return mockCmd("echo", "added 1 package")
	}

	r := update.UpdateResult{
		Tool: update.ToolInfo{
			Name:          "codegraph",
			InstallMethod: update.InstallNpmGlobal,
			NpmPackage:    "@colbymchenry/codegraph",
		},
		LatestVersion: "1.5.0",
		Status:        update.UpdateAvailable,
	}
	profile := system.PlatformProfile{OS: "linux", PackageManager: "apt"}

	_, err := runStrategy(context.Background(), r, profile)
	if err != nil {
		t.Fatalf("runStrategy npm-global: unexpected error: %v", err)
	}

	if gotName != "npm" {
		t.Errorf("exec name = %q, want %q", gotName, "npm")
	}
	wantArgs := []string{"install", "-g", "@colbymchenry/codegraph@latest"}
	if len(gotArgs) != len(wantArgs) {
		t.Fatalf("exec args = %v, want %v", gotArgs, wantArgs)
	}
	for i, want := range wantArgs {
		if gotArgs[i] != want {
			t.Errorf("exec args[%d] = %q, want %q (full args: %v)", i, gotArgs[i], want, gotArgs)
		}
	}
}

// TestRunStrategyNpmGlobalUpgradeError verifies npm failures propagate as
// errors with the command output attached.
func TestRunStrategyNpmGlobalUpgradeError(t *testing.T) {
	origExecCommand := execCommand
	t.Cleanup(func() { execCommand = origExecCommand })

	execCommand = func(name string, args ...string) *exec.Cmd {
		return mockCmd("false")
	}

	r := update.UpdateResult{
		Tool: update.ToolInfo{
			Name:          "codegraph",
			InstallMethod: update.InstallNpmGlobal,
			NpmPackage:    "@colbymchenry/codegraph",
		},
		LatestVersion: "1.5.0",
		Status:        update.UpdateAvailable,
	}
	profile := system.PlatformProfile{OS: "linux", PackageManager: "apt"}

	_, err := runStrategy(context.Background(), r, profile)
	if err == nil {
		t.Fatal("runStrategy npm-global: expected error when npm fails, got nil")
	}
	if !strings.Contains(err.Error(), "npm install -g @colbymchenry/codegraph@latest") {
		t.Errorf("error should name the failed command, got: %v", err)
	}
}

// TestRunStrategyNpmGlobalEmptyPackage verifies a missing NpmPackage degrades to
// a manual fallback instead of running a malformed npm command.
func TestRunStrategyNpmGlobalEmptyPackage(t *testing.T) {
	origExecCommand := execCommand
	t.Cleanup(func() { execCommand = origExecCommand })
	execCommand = func(name string, args ...string) *exec.Cmd {
		t.Fatalf("no command should run without an npm package, executed %s %v", name, args)
		return nil
	}

	r := update.UpdateResult{
		Tool: update.ToolInfo{
			Name:          "codegraph",
			InstallMethod: update.InstallNpmGlobal,
		},
		LatestVersion: "1.5.0",
		Status:        update.UpdateAvailable,
	}
	profile := system.PlatformProfile{OS: "linux", PackageManager: "apt"}

	_, err := runStrategy(context.Background(), r, profile)
	var fallback *ManualFallbackError
	if !errors.As(err, &fallback) {
		t.Fatalf("error = %v, want ManualFallbackError", err)
	}
}

// TestRunStrategyNpmGlobalSkipsHomebrewOwnership verifies npm-global tools never
// trigger Homebrew ownership detection, even on a brew profile: brew cannot own
// an npm package, and the detection would only add latency and failure modes.
func TestRunStrategyNpmGlobalSkipsHomebrewOwnership(t *testing.T) {
	origDetector := homebrewOwnershipDetector
	origExecCommand := execCommand
	t.Cleanup(func() {
		homebrewOwnershipDetector = origDetector
		execCommand = origExecCommand
	})

	homebrewOwnershipDetector = func(string) (update.HomebrewOwnership, error) {
		t.Fatal("homebrew ownership detection must not run for npm-global tools")
		return update.HomebrewNone, nil
	}
	execCommand = func(name string, args ...string) *exec.Cmd {
		return mockCmd("echo", "ok")
	}

	r := update.UpdateResult{
		Tool: update.ToolInfo{
			Name:          "codegraph",
			InstallMethod: update.InstallNpmGlobal,
			NpmPackage:    "@colbymchenry/codegraph",
		},
		LatestVersion: "1.5.0",
		Status:        update.UpdateAvailable,
	}
	profile := system.PlatformProfile{OS: "darwin", PackageManager: "brew"}

	if _, err := runStrategy(context.Background(), r, profile); err != nil {
		t.Fatalf("runStrategy npm-global on brew profile: unexpected error: %v", err)
	}
}
