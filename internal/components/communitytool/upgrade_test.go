package communitytool

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
)

func swapCodeGraphInstalledVersion(t *testing.T, version string) {
	t.Helper()
	previous := codeGraphInstalledVersion
	codeGraphInstalledVersion = func() string { return version }
	t.Cleanup(func() { codeGraphInstalledVersion = previous })
}

func TestUpgradeCodeGraphWithHome_CapturesAndReinstallsLatest(t *testing.T) {
	swapCodeGraphInstalledVersion(t, "1.4.1")
	home := t.TempDir()
	mustWrite(t, filepath.Join(home, ".claude", "settings.json"), `{}`)

	var commands []string
	result, err := UpgradeCodeGraphWithHome(home, "", RunnerFunc(func(name string, args ...string) error {
		command := strings.Join(append([]string{name}, args...), " ")
		commands = append(commands, command)
		if strings.HasPrefix(command, "codegraph install --target claude") {
			mustWrite(t, filepath.Join(home, ".claude.json"), `{"mcpServers":{"codegraph":{"command":"codegraph","args":["serve","--mcp"]}}}`)
		}
		return nil
	}), DetectorFunc(func(name string) (string, error) {
		if name == "codegraph" {
			return "/bin/codegraph", nil
		}
		return "", errors.New("not found")
	}))
	if err != nil {
		t.Fatalf("UpgradeCodeGraphWithHome() error = %v", err)
	}
	want := []string{
		"npm install -g @colbymchenry/codegraph@latest",
		"codegraph install --target claude --location global --yes",
	}
	if !reflect.DeepEqual(commands, want) {
		t.Fatalf("commands = %#v, want %#v", commands, want)
	}
	if !reflect.DeepEqual(result.CommandsRun, want) {
		t.Fatalf("result.CommandsRun = %#v, want %#v", result.CommandsRun, want)
	}
}

func TestInstallWithHomeOpts_ForceBypassesReconcileSatisfied(t *testing.T) {
	home := t.TempDir()
	mustWrite(t, filepath.Join(home, ".claude.json"), `{"mcpServers":{"codegraph":{"command":"codegraph"}}}`)
	mustWrite(t, filepath.Join(home, ".claude", "CLAUDE.md"), strings.Join([]string{
		"existing Claude guidance",
		"<!-- gentle-ai:codegraph-guidance -->",
		"CodeGraph guidance with `gentle-ai codegraph init --cwd <project-root>`",
		"<!-- /gentle-ai:codegraph-guidance -->",
	}, "\n"))

	detector := DetectorFunc(func(name string) (string, error) {
		if name == "codegraph" {
			return "/bin/codegraph", nil
		}
		return "", errors.New("not found")
	})

	// Force: a fully satisfied install must still run the global npm reinstall
	// and rewire the detected native targets.
	var forced []string
	if _, err := InstallWithHomeOpts(model.CommunityToolCodeGraph, "", home, RunnerFunc(func(name string, args ...string) error {
		forced = append(forced, strings.Join(append([]string{name}, args...), " "))
		return nil
	}), detector, InstallOpts{ForceReinstall: true}); err != nil {
		t.Fatalf("InstallWithHomeOpts(force) error = %v", err)
	}
	wantForced := []string{
		"npm install -g @colbymchenry/codegraph@latest",
		"codegraph install --target claude --location global --yes",
	}
	if !reflect.DeepEqual(forced, wantForced) {
		t.Fatalf("forced commands = %#v, want %#v", forced, wantForced)
	}

	// Triangulation: without force the same satisfied install stays untouched.
	var untouched []string
	if _, err := InstallWithHomeOpts(model.CommunityToolCodeGraph, "", home, RunnerFunc(func(name string, args ...string) error {
		untouched = append(untouched, strings.Join(append([]string{name}, args...), " "))
		return nil
	}), detector, InstallOpts{}); err != nil {
		t.Fatalf("InstallWithHomeOpts(zero) error = %v", err)
	}
	if len(untouched) != 0 {
		t.Fatalf("commands without force = %#v, want satisfied install untouched", untouched)
	}
}

func TestUpgradeCodeGraphWithHome_RollbackReinstallsCapturedVersion(t *testing.T) {
	swapCodeGraphInstalledVersion(t, "1.4.1")
	home := t.TempDir()
	mustWrite(t, filepath.Join(home, ".claude", "settings.json"), `{}`)
	guidancePath := filepath.Join(home, ".claude", "CLAUDE.md")
	mustWrite(t, guidancePath, "original guidance\n")

	var commands []string
	var targetInstalls int
	_, err := UpgradeCodeGraphWithHome(home, "", RunnerFunc(func(name string, args ...string) error {
		command := strings.Join(append([]string{name}, args...), " ")
		commands = append(commands, command)
		if strings.HasPrefix(command, "codegraph install --target claude") {
			targetInstalls++
			if targetInstalls == 1 {
				// The upgrade's rewire partially mutates managed files, then fails.
				mustWrite(t, filepath.Join(home, ".claude.json"), `{"mcpServers":{"codegraph":{"command":"codegraph"}}}`)
				mustWrite(t, guidancePath, "corrupted by failed upgrade\n")
				return errors.New("exit status 1")
			}
			mustWrite(t, filepath.Join(home, ".claude.json"), `{"mcpServers":{"codegraph":{"command":"codegraph","args":["serve","--mcp"]}}}`)
		}
		return nil
	}), DetectorFunc(func(name string) (string, error) {
		if name == "codegraph" {
			return "/bin/codegraph", nil
		}
		return "", errors.New("not found")
	}))
	if err == nil {
		t.Fatal("UpgradeCodeGraphWithHome() error = nil, want upgrade failure with successful rollback")
	}
	if !strings.Contains(err.Error(), "1.4.1") {
		t.Fatalf("error = %v, want captured version mentioned", err)
	}
	want := []string{
		"npm install -g @colbymchenry/codegraph@latest",
		"codegraph install --target claude --location global --yes",
		"npm install -g @colbymchenry/codegraph@1.4.1",
		"codegraph install --target claude --location global --yes",
	}
	if !reflect.DeepEqual(commands, want) {
		t.Fatalf("commands = %#v, want %#v", commands, want)
	}
	content, readErr := os.ReadFile(guidancePath)
	if readErr != nil {
		t.Fatalf("ReadFile(%q) error = %v", guidancePath, readErr)
	}
	if string(content) != "original guidance\n" {
		t.Fatalf("guidance file = %q, want snapshot restored", content)
	}
}

func TestUpgradeCodeGraphWithHome_RollbackFailureJoinsManualCommand(t *testing.T) {
	swapCodeGraphInstalledVersion(t, "1.4.1")
	home := t.TempDir()
	mustWrite(t, filepath.Join(home, ".claude", "settings.json"), `{}`)

	_, err := UpgradeCodeGraphWithHome(home, "", RunnerFunc(func(name string, args ...string) error {
		return errors.New("exit status 1")
	}), DetectorFunc(func(name string) (string, error) {
		if name == "codegraph" {
			return "/bin/codegraph", nil
		}
		return "", errors.New("not found")
	}))
	if err == nil {
		t.Fatal("UpgradeCodeGraphWithHome() error = nil, want joined upgrade+rollback failure")
	}
	for _, want := range []string{
		"npm install -g @colbymchenry/codegraph@1.4.1",
		"sync continues",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error = %v, want %q", err, want)
		}
	}
}
