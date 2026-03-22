package screens

import (
	"fmt"
	"strings"

	"github.com/gentleman-programming/gentle-ai/internal/backup"
	"github.com/gentleman-programming/gentle-ai/internal/tui/styles"
)

// RenderBackups renders the backup selection screen.
// It uses manifest.DisplayLabel() to show source + timestamp for each backup.
func RenderBackups(backups []backup.Manifest, cursor int) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Backup Management"))
	b.WriteString("\n\n")

	if len(backups) == 0 {
		b.WriteString(styles.WarningStyle.Render("No backups found yet."))
		b.WriteString("\n\n")
		b.WriteString(renderOptions([]string{"Back"}, 0))
		return b.String()
	}

	for idx, snapshot := range backups {
		// Use DisplayLabel for richer labels: "install — 2026-03-22 15:04 (5 files)"
		// Falls back to "unknown source — 2026-03-22 15:04" for old manifests.
		displayLabel := snapshot.DisplayLabel()
		if snapshot.CreatedByVersion != "" {
			displayLabel = fmt.Sprintf("%s  [v%s]", displayLabel, snapshot.CreatedByVersion)
		}
		label := fmt.Sprintf("%s  (%s)", snapshot.ID, displayLabel)
		focused := idx == cursor
		if focused {
			b.WriteString(styles.SelectedStyle.Render(styles.Cursor + label))
		} else {
			b.WriteString(styles.UnselectedStyle.Render("  " + label))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(renderOptions([]string{"Back"}, cursor-len(backups)))
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: select • esc: back"))

	return b.String()
}

// RenderRestoreConfirm renders the restore confirmation screen.
// It shows the backup identity and asks the user to confirm or cancel.
// Cursor 0 = "Restore", Cursor 1 = "Cancel".
func RenderRestoreConfirm(manifest backup.Manifest, cursor int) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Restore Backup"))
	b.WriteString("\n\n")

	b.WriteString(styles.HeadingStyle.Render("Backup: "))
	b.WriteString(styles.SelectedStyle.Render(manifest.ID))
	b.WriteString("\n")
	b.WriteString(styles.SubtextStyle.Render(manifest.DisplayLabel()))
	b.WriteString("\n\n")

	b.WriteString(styles.WarningStyle.Render("This will overwrite your current configuration."))
	b.WriteString("\n\n")

	b.WriteString(renderOptions([]string{"Restore", "Cancel"}, cursor))
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: select • esc: back"))

	return b.String()
}

// RenderRestoreResult renders the restore result screen.
// Shows a success message when err is nil, or an error message with details.
func RenderRestoreResult(manifest backup.Manifest, err error) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Restore Result"))
	b.WriteString("\n\n")

	if err == nil {
		b.WriteString(styles.SuccessStyle.Render("✓ Restore complete"))
		b.WriteString("\n\n")
		b.WriteString(styles.SubtextStyle.Render("Restored: "))
		b.WriteString(styles.SelectedStyle.Render(manifest.ID))
		b.WriteString("\n")
		b.WriteString(styles.SubtextStyle.Render(manifest.DisplayLabel()))
		b.WriteString("\n\n")
		b.WriteString(styles.UnselectedStyle.Render("Your configuration has been restored from this backup."))
	} else {
		b.WriteString(styles.ErrorStyle.Render("✗ Restore failed"))
		b.WriteString("\n\n")
		b.WriteString(styles.SubtextStyle.Render("Backup: "))
		b.WriteString(styles.SelectedStyle.Render(manifest.ID))
		b.WriteString("\n\n")
		b.WriteString(styles.HeadingStyle.Render("Error:"))
		b.WriteString("\n")
		b.WriteString(styles.ErrorStyle.Render("  " + err.Error()))
		b.WriteString("\n\n")
		b.WriteString(styles.SubtextStyle.Render("Your files were not modified."))
	}

	b.WriteString("\n\n")
	b.WriteString(styles.HelpStyle.Render("enter: back to backups • esc: back"))

	return b.String()
}
