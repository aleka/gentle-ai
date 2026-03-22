// Package upgrade provides the upgrade executor for managed tools.
// It sits ON TOP of the read-only internal/update package and is deliberately
// isolated from install, pipeline, planner, and config-sync code paths.
//
// Import boundary: this package MUST NOT import:
//   - github.com/gentleman-programming/gentle-ai/internal/pipeline
//   - github.com/gentleman-programming/gentle-ai/internal/planner
//   - github.com/gentleman-programming/gentle-ai/internal/cli
package upgrade

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gentleman-programming/gentle-ai/internal/backup"
	"github.com/gentleman-programming/gentle-ai/internal/system"
	"github.com/gentleman-programming/gentle-ai/internal/update"
)

// Package-level vars for testability — same pattern as internal/update/detect.go.
// execCommand is used as: execCommand(name, args...) — identical signature to exec.Command.
// Swapping this var in tests controls which commands are actually run.
var execCommand = exec.Command

// configPathsForBackup returns the well-known agent config file paths that the
// backup snapshot must include before any upgrade execution.
// These are the same paths scanned by system.ScanConfigs.
func configPathsForBackup(homeDir string) []string {
	return []string{
		filepath.Join(homeDir, ".claude", "CLAUDE.md"),
		filepath.Join(homeDir, ".config", "opencode", "config.json"),
		filepath.Join(homeDir, ".gemini", "GEMINI.md"),
		filepath.Join(homeDir, ".cursor", "rules"),
	}
}

// Execute evaluates UpdateResults, snapshots config before execution, then runs
// the appropriate upgrade strategy for each eligible tool.
//
// Reporting rules:
//   - Status UpdateAvailable → attempt upgrade; report Succeeded/Failed/Skipped(manual)
//   - Status DevBuild → report as UpgradeSkipped with ManualHint (dev/source build)
//   - Status UpToDate, NotInstalled, CheckFailed, VersionUnknown → omitted from report
//   - dryRun=true → no exec; eligible tools reported as UpgradeSkipped
//
// The backup snapshot is created before any exec call — this is the architectural
// guarantee that config is safe even if an upgrade fails mid-way.
func Execute(ctx context.Context, results []update.UpdateResult, profile system.PlatformProfile, homeDir string, dryRun bool) UpgradeReport {
	// Separate tools into executable (UpdateAvailable) and dev-build (DevBuild).
	// DevBuild tools are included in the report as UpgradeSkipped with a clear hint.
	var executable []update.UpdateResult
	var devBuilds []update.UpdateResult
	for _, r := range results {
		switch r.Status {
		case update.UpdateAvailable:
			executable = append(executable, r)
		case update.DevBuild:
			devBuilds = append(devBuilds, r)
			// UpToDate, NotInstalled, CheckFailed, VersionUnknown → omit from report
		}
	}

	// If nothing is executable or dev-built, return empty report.
	if len(executable) == 0 && len(devBuilds) == 0 {
		return UpgradeReport{DryRun: dryRun}
	}

	// Create backup snapshot BEFORE any execution (only when there are executables).
	backupID := ""
	if !dryRun && len(executable) > 0 {
		snapshotDir := filepath.Join(homeDir, ".gentle-ai", "backups",
			fmt.Sprintf("upgrade-%s", time.Now().UTC().Format("20060102T150405Z")))
		snap := backup.NewSnapshotter()
		manifest, err := snap.Create(snapshotDir, configPathsForBackup(homeDir))
		if err == nil {
			backupID = manifest.ID
		}
		// Non-fatal: if backup fails we still proceed and set BackupID empty.
	}

	// Build results slice: dev-build skips first (no exec), then executable tools.
	toolResults := make([]ToolUpgradeResult, 0, len(executable)+len(devBuilds))

	// Dev-build tools: always UpgradeSkipped with a source-build hint.
	for _, r := range devBuilds {
		toolResults = append(toolResults, ToolUpgradeResult{
			ToolName:   r.Tool.Name,
			OldVersion: r.InstalledVersion,
			NewVersion: r.LatestVersion,
			Method:     effectiveMethod(r.Tool, profile),
			Status:     UpgradeSkipped,
			ManualHint: fmt.Sprintf("source build — upgrade manually or install a release binary from https://github.com/Gentleman-Programming/%s/releases", r.Tool.Repo),
		})
	}

	// Executable tools: run upgrade strategy.
	for _, r := range executable {
		toolResult := executeOne(ctx, r, profile, dryRun)
		toolResults = append(toolResults, toolResult)
	}

	return UpgradeReport{
		BackupID: backupID,
		Results:  toolResults,
		DryRun:   dryRun,
	}
}

// executeOne runs the upgrade for a single tool.
func executeOne(ctx context.Context, r update.UpdateResult, profile system.PlatformProfile, dryRun bool) ToolUpgradeResult {
	base := ToolUpgradeResult{
		ToolName:   r.Tool.Name,
		OldVersion: r.InstalledVersion,
		NewVersion: r.LatestVersion,
		Method:     effectiveMethod(r.Tool, profile),
	}

	if dryRun {
		base.Status = UpgradeSkipped
		return base
	}

	err := runStrategy(ctx, r, profile)
	if err != nil {
		// Distinguish manual fallback (informational skip) from real failures.
		if hint, ok := AsManualFallback(err); ok {
			base.Status = UpgradeSkipped
			base.ManualHint = hint
			// Err is intentionally nil: a manual skip is not an error condition.
		} else {
			base.Status = UpgradeFailed
			base.Err = err
		}
	} else {
		base.Status = UpgradeSucceeded
	}

	return base
}

// effectiveMethod resolves the actual upgrade strategy for a tool on a given platform.
// On brew-managed platforms, brew takes precedence over the tool's declared method.
func effectiveMethod(tool update.ToolInfo, profile system.PlatformProfile) update.InstallMethod {
	if profile.PackageManager == "brew" {
		return update.InstallBrew
	}
	return tool.InstallMethod
}
