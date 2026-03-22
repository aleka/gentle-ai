package screens

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gentleman-programming/gentle-ai/internal/backup"
)

// TestRenderBackupsShowsDisplayLabel verifies that RenderBackups uses the
// manifest's DisplayLabel (source + timestamp) instead of just the raw ID.
func TestRenderBackupsShowsDisplayLabel(t *testing.T) {
	manifests := []backup.Manifest{
		{
			ID:        "20260322150405.000000000",
			CreatedAt: time.Date(2026, 3, 22, 15, 4, 5, 0, time.UTC),
			Source:    backup.BackupSourceInstall,
		},
	}

	output := RenderBackups(manifests, 0)

	// Must include the source label from DisplayLabel.
	if !strings.Contains(output, "install") {
		t.Errorf("RenderBackups should show source label 'install' from DisplayLabel; got:\n%s", output)
	}
}

// TestRenderBackupsShowsFallbackLabelForOldManifest verifies that an old
// manifest without Source metadata renders with "unknown source" fallback.
func TestRenderBackupsShowsFallbackLabelForOldManifest(t *testing.T) {
	manifests := []backup.Manifest{
		{
			ID:        "old-backup-id",
			CreatedAt: time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC),
			// Source intentionally empty — simulates old manifest.
		},
	}

	output := RenderBackups(manifests, 0)

	if !strings.Contains(output, "unknown source") {
		t.Errorf("RenderBackups should show 'unknown source' for old manifests; got:\n%s", output)
	}
}

// TestRenderRestoreConfirmIncludesBackupIdentity verifies the confirm screen
// shows the backup ID and source label so the user knows what they're restoring.
func TestRenderRestoreConfirmIncludesBackupIdentity(t *testing.T) {
	manifest := backup.Manifest{
		ID:        "20260322150405.000000000",
		CreatedAt: time.Date(2026, 3, 22, 15, 4, 5, 0, time.UTC),
		Source:    backup.BackupSourceSync,
	}

	output := RenderRestoreConfirm(manifest, 0)

	if !strings.Contains(output, manifest.ID) {
		t.Errorf("RenderRestoreConfirm should show backup ID; got:\n%s", output)
	}

	if !strings.Contains(output, "sync") {
		t.Errorf("RenderRestoreConfirm should show source label; got:\n%s", output)
	}
}

// TestRenderRestoreConfirmShowsConfirmAndCancelOptions verifies that the
// confirmation screen presents both "Restore" and "Cancel" options.
func TestRenderRestoreConfirmShowsConfirmAndCancelOptions(t *testing.T) {
	manifest := backup.Manifest{
		ID:        "test-backup",
		CreatedAt: time.Now().UTC(),
	}

	output := RenderRestoreConfirm(manifest, 0)

	// Must show a restore/confirm action.
	if !strings.Contains(strings.ToLower(output), "restore") {
		t.Errorf("RenderRestoreConfirm missing restore option; got:\n%s", output)
	}

	// Must show a cancel action.
	if !strings.Contains(strings.ToLower(output), "cancel") && !strings.Contains(strings.ToLower(output), "back") {
		t.Errorf("RenderRestoreConfirm missing cancel/back option; got:\n%s", output)
	}
}

// TestRenderRestoreResultSuccessShowsSuccessMessage verifies that a successful
// restore result screen displays a success confirmation.
func TestRenderRestoreResultSuccessShowsSuccessMessage(t *testing.T) {
	manifest := backup.Manifest{
		ID:        "my-backup-001",
		CreatedAt: time.Now().UTC(),
		Source:    backup.BackupSourceUpgrade,
	}

	output := RenderRestoreResult(manifest, nil)

	// Must include a success indicator.
	lower := strings.ToLower(output)
	if !strings.Contains(lower, "success") && !strings.Contains(lower, "restored") && !strings.Contains(lower, "complete") {
		t.Errorf("RenderRestoreResult(nil err) should show success; got:\n%s", output)
	}

	// Must show the backup identity.
	if !strings.Contains(output, manifest.ID) {
		t.Errorf("RenderRestoreResult should show backup ID; got:\n%s", output)
	}
}

// TestRenderRestoreResultFailureShowsErrorMessage verifies that a failed
// restore result screen displays actionable failure text.
func TestRenderRestoreResultFailureShowsErrorMessage(t *testing.T) {
	manifest := backup.Manifest{
		ID:        "my-backup-002",
		CreatedAt: time.Now().UTC(),
	}

	errText := "snapshot file missing"
	output := RenderRestoreResult(manifest, fmt.Errorf("%s", errText))

	lower := strings.ToLower(output)
	if !strings.Contains(lower, "fail") && !strings.Contains(lower, "error") {
		t.Errorf("RenderRestoreResult(err) should show failure; got:\n%s", output)
	}

	if !strings.Contains(output, errText) {
		t.Errorf("RenderRestoreResult should include error text %q; got:\n%s", errText, output)
	}
}
