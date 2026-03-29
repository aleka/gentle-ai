package state

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const stateDir = ".gentle-ai"
const stateFile = "state.json"

// InstallState holds the persisted user selections from the last install run.
type InstallState struct {
	InstalledAgents []string `json:"installed_agents"`
}

// Path returns the absolute path to the state file for the given home directory.
func Path(homeDir string) string {
	return filepath.Join(homeDir, stateDir, stateFile)
}

// Read reads and unmarshals the state file from the given home directory.
// Returns an error if the file does not exist or cannot be decoded.
func Read(homeDir string) (InstallState, error) {
	data, err := os.ReadFile(Path(homeDir))
	if err != nil {
		return InstallState{}, err
	}
	var s InstallState
	if err := json.Unmarshal(data, &s); err != nil {
		return InstallState{}, err
	}
	return s, nil
}

// InstalledAgentsFromState returns the persisted agent names from state.json.
// When the file exists and contains at least one installed agent, the names
// are returned. Returns nil when the file is absent, unreadable, or empty —
// callers should fall back to filesystem detection in that case.
func InstalledAgentsFromState(homeDir string) []string {
	if homeDir == "" {
		return nil
	}
	s, err := Read(homeDir)
	if err != nil || len(s.InstalledAgents) == 0 {
		return nil
	}
	return s.InstalledAgents
}

// Write persists the given agent IDs to the state file under the given home directory.
// It creates the .gentle-ai directory if it does not already exist.
func Write(homeDir string, agents []string) error {
	dir := filepath.Join(homeDir, stateDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	s := InstallState{InstalledAgents: agents}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(homeDir), append(data, '\n'), 0o644)
}
