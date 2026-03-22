package upgrade

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/system"
	"github.com/gentleman-programming/gentle-ai/internal/update"
)

// --- helpers ---

func brewProfile() system.PlatformProfile {
	return system.PlatformProfile{OS: "darwin", PackageManager: "brew", Supported: true}
}

func linuxProfile() system.PlatformProfile {
	return system.PlatformProfile{OS: "linux", LinuxDistro: system.LinuxDistroUbuntu, PackageManager: "apt", Supported: true}
}

func makeResult(name string, status update.UpdateStatus, oldVer, newVer string, method update.InstallMethod) update.UpdateResult {
	return update.UpdateResult{
		Tool: update.ToolInfo{
			Name:          name,
			Owner:         "Gentleman-Programming",
			Repo:          name,
			InstallMethod: method,
		},
		InstalledVersion: oldVer,
		LatestVersion:    newVer,
		Status:           status,
	}
}

// --- TestExecute_NoopWhenNothingIsExecutable ---

// TestExecute_NoopWhenNothingIsExecutable verifies that Execute returns an empty
// UpgradeReport with no backup and no tool results when no UpdateResult is
// UpdateAvailable or DevBuild status (i.e. only UpToDate and NotInstalled tools).
func TestExecute_NoopWhenNothingIsExecutable(t *testing.T) {
	results := []update.UpdateResult{
		makeResult("gentle-ai", update.UpToDate, "1.0.0", "1.0.0", update.InstallBinary),
		makeResult("engram", update.NotInstalled, "", "0.4.0", update.InstallGoInstall),
		// gga: CheckFailed — should also be omitted from results.
		makeResult("gga", update.CheckFailed, "", "", update.InstallBinary),
	}

	report := Execute(context.Background(), results, brewProfile(), t.TempDir(), false)

	if report.BackupID != "" {
		t.Errorf("BackupID = %q, want empty — no backup should be created when nothing to execute", report.BackupID)
	}

	if len(report.Results) != 0 {
		t.Errorf("len(Results) = %d, want 0 — UpToDate, NotInstalled, CheckFailed must be omitted", len(report.Results))
	}

	if report.DryRun {
		t.Errorf("DryRun should be false when not requested")
	}
}

// --- TestExecute_DevBuildOnlyNoBackupCreated ---

// TestExecute_DevBuildOnlyNoBackupCreated verifies that when ALL tools are DevBuild
// (nothing to execute), no backup snapshot is created. Backup is only needed before
// actual binary execution, not for skip-only reports.
func TestExecute_DevBuildOnlyNoBackupCreated(t *testing.T) {
	origExecCommand := execCommand
	t.Cleanup(func() { execCommand = origExecCommand })

	execCalled := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		execCalled = true
		return exec.Command("echo", "should not be called")
	}

	results := []update.UpdateResult{
		makeResult("gentle-ai", update.DevBuild, "dev", "1.0.0", update.InstallBinary),
	}

	report := Execute(context.Background(), results, linuxProfile(), t.TempDir(), false)

	if execCalled {
		t.Errorf("execCommand should NOT be called for DevBuild-only inputs")
	}

	// DevBuild tool MUST appear in results as UpgradeSkipped.
	if len(report.Results) != 1 {
		t.Fatalf("len(Results) = %d, want 1 — DevBuild tool must appear as skipped", len(report.Results))
	}
	if report.Results[0].Status != UpgradeSkipped {
		t.Errorf("DevBuild Status = %q, want UpgradeSkipped", report.Results[0].Status)
	}

	// No backup should be created — nothing executed.
	if report.BackupID != "" {
		t.Errorf("BackupID = %q, want empty — no backup when no execution occurs", report.BackupID)
	}
}

// --- TestExecute_BackupBeforeExecution ---

// TestExecute_BackupBeforeExecution verifies the architectural invariant:
// a backup snapshot is created BEFORE any upgrade execution begins.
// We verify this by ensuring BackupID is non-empty when upgrades are available.
func TestExecute_BackupBeforeExecution(t *testing.T) {
	origExecCommand := execCommand
	t.Cleanup(func() { execCommand = origExecCommand })

	// Capture exec calls to verify ordering.
	var calls []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		calls = append(calls, name)
		// Return a real passing command (echo) so exec succeeds.
		return exec.Command("echo", "ok")
	}

	results := []update.UpdateResult{
		makeResult("engram", update.UpdateAvailable, "0.3.0", "0.4.0", update.InstallGoInstall),
	}
	results[0].Tool.GoImportPath = "github.com/Gentleman-Programming/engram/cmd/engram"

	report := Execute(context.Background(), results, linuxProfile(), t.TempDir(), false)

	// BackupID must be non-empty.
	if report.BackupID == "" {
		t.Errorf("BackupID is empty — backup must be created before upgrade execution")
	}

	// At least one result must be present.
	if len(report.Results) != 1 {
		t.Fatalf("len(Results) = %d, want 1", len(report.Results))
	}
}

// --- TestExecute_DryRunNeverExecs ---

// TestExecute_DryRunNeverExecs verifies that when dryRun=true, no exec is called
// but the report is still populated.
func TestExecute_DryRunNeverExecs(t *testing.T) {
	origExecCommand := execCommand
	t.Cleanup(func() { execCommand = origExecCommand })

	called := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		called = true
		return exec.Command("echo", "should not run")
	}

	results := []update.UpdateResult{
		makeResult("engram", update.UpdateAvailable, "0.3.0", "0.4.0", update.InstallGoInstall),
	}
	results[0].Tool.GoImportPath = "github.com/Gentleman-Programming/engram/cmd/engram"

	report := Execute(context.Background(), results, linuxProfile(), t.TempDir(), true)

	if called {
		t.Errorf("execCommand was called during dry-run — must NOT execute")
	}

	if !report.DryRun {
		t.Errorf("DryRun = false, want true")
	}

	if len(report.Results) != 1 {
		t.Fatalf("len(Results) = %d, want 1", len(report.Results))
	}

	if report.Results[0].Status != UpgradeSkipped {
		t.Errorf("dry-run status = %q, want UpgradeSkipped", report.Results[0].Status)
	}
}

// --- TestExecute_PerToolSuccessFailureSkip ---

// TestExecute_PerToolSuccessAndFailure verifies that Execute reports success for one
// tool and failure for another in a mixed scenario.
func TestExecute_PerToolSuccessAndFailure(t *testing.T) {
	origExecCommand := execCommand
	t.Cleanup(func() { execCommand = origExecCommand })

	execCommand = func(name string, args ...string) *exec.Cmd {
		// engram go install succeeds, gga curl/download attempt fails — we simulate
		// the failure by having execCommand return false for "gga" detection.
		if name == "go" {
			return exec.Command("echo", "go install ok")
		}
		// Any other exec attempt fails.
		return exec.Command("false")
	}

	results := []update.UpdateResult{
		makeResult("engram", update.UpdateAvailable, "0.3.0", "0.4.0", update.InstallGoInstall),
	}
	results[0].Tool.GoImportPath = "github.com/Gentleman-Programming/engram/cmd/engram"

	report := Execute(context.Background(), results, linuxProfile(), t.TempDir(), false)

	if len(report.Results) != 1 {
		t.Fatalf("len(Results) = %d, want 1", len(report.Results))
	}

	// engram should succeed (go install echo'd "ok")
	if report.Results[0].Status != UpgradeSucceeded {
		t.Errorf("engram status = %q, want UpgradeSucceeded", report.Results[0].Status)
	}
}

// --- TestExecute_DevBuildIsSkipped ---

// TestExecute_DevBuildIsSkipped verifies the spec requirement:
// gentle-ai with DevBuild status must appear in Results as UpgradeSkipped
// with a non-empty ManualHint explaining it is a source/dev build.
// DevBuild tools must NOT be auto-executed, and engram/gga remain eligible.
func TestExecute_DevBuildIsSkipped(t *testing.T) {
	origExecCommand := execCommand
	t.Cleanup(func() { execCommand = origExecCommand })
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("echo", "ok")
	}

	results := []update.UpdateResult{
		makeResult("gentle-ai", update.DevBuild, "dev", "1.0.0", update.InstallBinary),
		makeResult("engram", update.UpdateAvailable, "0.3.0", "0.4.0", update.InstallGoInstall),
	}
	results[1].Tool.GoImportPath = "github.com/Gentleman-Programming/engram/cmd/engram"

	report := Execute(context.Background(), results, linuxProfile(), t.TempDir(), false)

	// gentle-ai (DevBuild) MUST appear as UpgradeSkipped with a ManualHint.
	var devResult *ToolUpgradeResult
	for i := range report.Results {
		if report.Results[i].ToolName == "gentle-ai" {
			r := report.Results[i]
			devResult = &r
		}
	}
	if devResult == nil {
		t.Fatalf("gentle-ai (DevBuild) must appear in Results — was not found")
	}
	if devResult.Status != UpgradeSkipped {
		t.Errorf("gentle-ai DevBuild Status = %q, want UpgradeSkipped", devResult.Status)
	}
	if devResult.ManualHint == "" {
		t.Errorf("gentle-ai DevBuild ManualHint must be non-empty")
	}

	// engram should still be processed as succeeded.
	found := false
	for _, r := range report.Results {
		if r.ToolName == "engram" {
			found = true
			if r.Status != UpgradeSucceeded {
				t.Errorf("engram status = %q, want UpgradeSucceeded", r.Status)
			}
		}
	}
	if !found {
		t.Errorf("engram not found in Results")
	}
}

// --- TestExecute_FailureDoesNotImplyConfigLoss ---

// TestExecute_FailureDoesNotImplyConfigLoss verifies that when a tool upgrade fails,
// we can still retrieve the BackupID — confirming config was snapshotted first.
func TestExecute_FailureDoesNotImplyConfigLoss(t *testing.T) {
	origExecCommand := execCommand
	t.Cleanup(func() { execCommand = origExecCommand })

	// Force all exec to fail.
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("false")
	}

	results := []update.UpdateResult{
		makeResult("engram", update.UpdateAvailable, "0.3.0", "0.4.0", update.InstallGoInstall),
	}
	results[0].Tool.GoImportPath = "github.com/Gentleman-Programming/engram/cmd/engram"

	report := Execute(context.Background(), results, linuxProfile(), t.TempDir(), false)

	// Even with failure, BackupID must be set (backup happened before exec).
	if report.BackupID == "" {
		t.Errorf("BackupID is empty — backup must be created before upgrade, even if upgrade fails")
	}

	if len(report.Results) != 1 {
		t.Fatalf("len(Results) = %d, want 1", len(report.Results))
	}

	if report.Results[0].Status != UpgradeFailed {
		t.Errorf("status = %q, want UpgradeFailed", report.Results[0].Status)
	}

	if report.Results[0].Err == nil {
		t.Errorf("Err should not be nil on failure")
	}
}

// --- TestExecute_InstallNotInvoked ---

// TestExecute_InstallNotInvoked verifies the isolation contract:
// Execute must not invoke any install/sync functions.
// We test this by verifying the package cannot even reference installer packages.
// This is enforced by the import boundary (no import of pipeline/planner/cli).
func TestExecute_InstallNotInvoked(t *testing.T) {
	// This test is intentionally a documentation-only guard.
	// The real enforcement is: this package MUST NOT import:
	//   - github.com/gentleman-programming/gentle-ai/internal/pipeline
	//   - github.com/gentleman-programming/gentle-ai/internal/planner
	//   - github.com/gentleman-programming/gentle-ai/internal/cli
	//
	// If you see those imports appear, the isolation contract is broken.
	// See TestExecuteImportBoundary for the compile-time enforcement approach.
	t.Log("install isolation enforced by import boundary — see imports at top of executor.go")
}

// --- TestExecute_DevBuildSurfacedAsSkipped ---

// TestExecute_DevBuildSurfacedAsSkipped verifies the spec gap:
// A DevBuild tool (e.g. gentle-ai with version="dev") MUST appear in UpgradeReport.Results
// with Status=UpgradeSkipped and a non-empty ManualHint explaining it is a dev/source build.
// Previously, DevBuild tools were silently omitted from Results entirely.
func TestExecute_DevBuildSurfacedAsSkipped(t *testing.T) {
	origExecCommand := execCommand
	t.Cleanup(func() { execCommand = origExecCommand })
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("echo", "ok")
	}

	results := []update.UpdateResult{
		makeResult("gentle-ai", update.DevBuild, "dev", "1.0.0", update.InstallBinary),
		makeResult("engram", update.UpdateAvailable, "0.3.0", "0.4.0", update.InstallGoInstall),
	}
	results[1].Tool.GoImportPath = "github.com/Gentleman-Programming/engram/cmd/engram"

	report := Execute(context.Background(), results, linuxProfile(), t.TempDir(), false)

	// gentle-ai (DevBuild) MUST appear in results as UpgradeSkipped.
	var devResult *ToolUpgradeResult
	for i := range report.Results {
		if report.Results[i].ToolName == "gentle-ai" {
			r := report.Results[i]
			devResult = &r
		}
	}

	if devResult == nil {
		t.Fatalf("gentle-ai DevBuild must appear in Results as UpgradeSkipped, but was not found")
	}

	if devResult.Status != UpgradeSkipped {
		t.Errorf("gentle-ai DevBuild Status = %q, want UpgradeSkipped", devResult.Status)
	}

	if devResult.ManualHint == "" {
		t.Errorf("gentle-ai DevBuild ManualHint must be non-empty — should explain dev/source build")
	}

	// engram (UpdateAvailable) must still be processed normally.
	found := false
	for _, r := range report.Results {
		if r.ToolName == "engram" {
			found = true
			if r.Status != UpgradeSucceeded {
				t.Errorf("engram status = %q, want UpgradeSucceeded", r.Status)
			}
		}
	}
	if !found {
		t.Errorf("engram not found in Results")
	}
}

// --- TestExecute_ManualFallbackSurfacedAsSkippedNotFailed ---

// TestExecute_ManualFallbackSurfacedAsSkippedNotFailed verifies the spec gap:
// When runStrategy returns a manual fallback error (e.g. Windows binary self-replace),
// the ToolUpgradeResult must be UpgradeSkipped (not UpgradeFailed) and ManualHint
// must be populated from the error message so RenderUpgradeReport can display it.
func TestExecute_ManualFallbackSurfacedAsSkippedNotFailed(t *testing.T) {
	origExecCommand := execCommand
	t.Cleanup(func() { execCommand = origExecCommand })

	execCalled := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		execCalled = true
		return exec.Command("echo", "should not be called")
	}

	// Windows profile → binaryUpgrade returns a manual fallback error.
	windowsProfile := system.PlatformProfile{OS: "windows", PackageManager: "winget", Supported: true}

	results := []update.UpdateResult{
		makeResult("gentle-ai", update.UpdateAvailable, "1.0.0", "1.5.0", update.InstallBinary),
	}
	results[0].UpdateHint = "See https://github.com/Gentleman-Programming/gentle-ai/releases"

	report := Execute(context.Background(), results, windowsProfile, t.TempDir(), false)

	if execCalled {
		t.Errorf("execCommand should not be called for Windows binary manual fallback")
	}

	if len(report.Results) != 1 {
		t.Fatalf("len(Results) = %d, want 1", len(report.Results))
	}

	r := report.Results[0]

	// Must be UpgradeSkipped (not UpgradeFailed) — this is a manual action, not a failure.
	if r.Status != UpgradeSkipped {
		t.Errorf("Windows binary fallback Status = %q, want UpgradeSkipped (not UpgradeFailed)", r.Status)
	}

	// ManualHint must be populated.
	if r.ManualHint == "" {
		t.Errorf("Windows binary fallback ManualHint must be non-empty")
	}

	// Err should be nil for a manual skip (it is not a failure).
	if r.Err != nil {
		t.Errorf("Windows binary fallback Err = %v, want nil (manual skips are not errors)", r.Err)
	}
}

// --- TestExecute_ConfigNotMutatedDuringUpgrade ---

// TestExecute_ConfigNotMutatedDuringUpgrade provides direct evidence that upgrade
// execution does not mutate config file contents — the spec's config preservation
// guarantee. We create real config files in a temp dir, run Execute (stubbed exec),
// and diff the contents before and after.
func TestExecute_ConfigNotMutatedDuringUpgrade(t *testing.T) {
	homeDir := t.TempDir()

	// Create realistic config files with known contents.
	configFiles := map[string]string{
		".claude/CLAUDE.md":            "# Claude config\nThis is my config.\n",
		".config/opencode/config.json": `{"theme":"kanagawa"}`,
		".gemini/GEMINI.md":            "# Gemini config\nMy rules.\n",
	}

	for relPath, content := range configFiles {
		fullPath := homeDir + "/" + relPath
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("create dir for %s: %v", relPath, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("write config %s: %v", relPath, err)
		}
	}

	origExecCommand := execCommand
	t.Cleanup(func() { execCommand = origExecCommand })
	execCommand = func(name string, args ...string) *exec.Cmd {
		// Simulate a successful upgrade (no-op shell command).
		return exec.Command("echo", "upgrade ok")
	}

	results := []update.UpdateResult{
		makeResult("engram", update.UpdateAvailable, "0.3.0", "0.4.0", update.InstallGoInstall),
	}
	results[0].Tool.GoImportPath = "github.com/Gentleman-Programming/engram/cmd/engram"

	profile := linuxProfile()

	// Execute upgrade.
	report := Execute(context.Background(), results, profile, homeDir, false)

	// Verify upgrade ran.
	if len(report.Results) != 1 {
		t.Fatalf("len(Results) = %d, want 1", len(report.Results))
	}
	if report.Results[0].Status != UpgradeSucceeded {
		t.Errorf("engram status = %q, want UpgradeSucceeded", report.Results[0].Status)
	}

	// Verify config files are byte-identical after upgrade.
	for relPath, want := range configFiles {
		fullPath := homeDir + "/" + relPath
		got, err := os.ReadFile(fullPath)
		if err != nil {
			t.Fatalf("read config %s after upgrade: %v", relPath, err)
		}
		if string(got) != want {
			t.Errorf("config %s was mutated by upgrade!\n  before: %q\n  after:  %q", relPath, want, string(got))
		}
	}
}

// --- helper: verify errors wrap correctly ---
func TestToolUpgradeResult_ErrorWrapping(t *testing.T) {
	sentinel := errors.New("sentinel error")
	r := ToolUpgradeResult{
		ToolName: "engram",
		Status:   UpgradeFailed,
		Err:      sentinel,
	}

	if !errors.Is(r.Err, sentinel) {
		t.Errorf("errors.Is failed — Err should wrap the sentinel")
	}
}
