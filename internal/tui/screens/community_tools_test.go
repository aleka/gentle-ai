package screens

import (
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/components/communitytool"
	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
	"github.com/gentleman-programming/gentle-ai/v2/internal/update"
)

func TestRenderCommunityToolsShowsCodeGraph(t *testing.T) {
	out := RenderCommunityTools([]model.CommunityToolID{model.CommunityToolCodeGraph}, 0, nil, false, nil, nil)
	for _, want := range []string{"Community Tools/Plugins", "[x] CodeGraph", "View repo: https://github.com/colbymchenry/codegraph", "Continue", "Back"} {
		if !strings.Contains(out, want) {
			t.Fatalf("RenderCommunityTools missing %q; output:\n%s", want, out)
		}
	}
}

func TestRenderCommunityToolsShowsStatusLoadingAndAgentState(t *testing.T) {
	loading := RenderCommunityTools(nil, 0, nil, true, nil, nil)
	if !strings.Contains(loading, "Detecting installed tool and agent wiring") {
		t.Fatalf("loading output missing detection text:\n%s", loading)
	}

	status := communitytool.Status{
		Tool: model.CommunityToolCodeGraph,
		CLI:  communitytool.AvailabilityAvailable,
		Agents: []communitytool.AgentStatus{
			{Agent: model.AgentClaudeCode, Name: "Claude Code", Detected: true, Configured: true, Status: communitytool.AgentStatusConfigured, Path: "/tmp/.claude/mcp/codegraph.json"},
			{Agent: model.AgentOpenCode, Name: "OpenCode", Detected: true, Configured: false, Status: communitytool.AgentStatusMissing, Path: "/tmp/.config/opencode"},
		},
	}
	out := RenderCommunityTools(nil, 0, []communitytool.Status{status}, false, nil, nil)
	for _, want := range []string{"CodeGraph CLI: available", "Agent wiring: 2 detected • 1 configured • 1 missing", "Claude Code: configured", "OpenCode: missing"} {
		if !strings.Contains(out, want) {
			t.Fatalf("status output missing %q; output:\n%s", want, out)
		}
	}
}

// TestRenderCommunityToolsShowsVersionPair verifies tasks 4.7/4.8: when the
// update results carry a CodeGraph update hint, the screen surfaces the
// installed → latest version pair (spec SHOULD: Community Tools screen shows
// the same pair as the upgrade screen and banner).
func TestRenderCommunityToolsShowsVersionPair(t *testing.T) {
	codeGraphUpdate := update.UpdateResult{
		Tool:             update.ToolInfo{Name: "codegraph"},
		Status:           update.UpdateAvailable,
		InstalledVersion: "1.4.1",
		LatestVersion:    "1.5.0",
	}

	tests := []struct {
		name    string
		updates []update.UpdateResult
		want    string
		notWant string
	}{
		{
			name:    "codegraph update available shows version pair",
			updates: []update.UpdateResult{codeGraphUpdate},
			want:    "CodeGraph 1.4.1 → 1.5.0",
		},
		{
			name: "codegraph up to date shows no pair",
			updates: []update.UpdateResult{
				{Tool: update.ToolInfo{Name: "codegraph"}, Status: update.UpToDate, InstalledVersion: "1.5.0", LatestVersion: "1.5.0"},
			},
			notWant: "1.4.1 →",
		},
		{
			name: "other tool update shows no codegraph pair",
			updates: []update.UpdateResult{
				{Tool: update.ToolInfo{Name: "engram"}, Status: update.UpdateAvailable, InstalledVersion: "0.3.2", LatestVersion: "0.4.0"},
			},
			notWant: "CodeGraph 0.3.2",
		},
		{
			name:    "nil update results show no pair",
			updates: nil,
			notWant: "→",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := RenderCommunityTools([]model.CommunityToolID{model.CommunityToolCodeGraph}, 0, nil, false, nil, tt.updates)
			if tt.want != "" && !strings.Contains(out, tt.want) {
				t.Fatalf("output missing version pair %q; output:\n%s", tt.want, out)
			}
			if tt.notWant != "" && strings.Contains(out, tt.notWant) {
				t.Fatalf("output must not contain %q; output:\n%s", tt.notWant, out)
			}
		})
	}
}

func TestRenderCommunityToolResultShowsPartialContextOnError(t *testing.T) {
	result := communitytool.Result{
		Tool: model.CommunityToolCodeGraph,
		StatusAfter: &communitytool.Status{
			Tool: model.CommunityToolCodeGraph,
			CLI:  communitytool.AvailabilityMissing,
			Agents: []communitytool.AgentStatus{
				{Agent: model.AgentOpenCode, Name: "OpenCode", Detected: true, Configured: false, Status: communitytool.AgentStatusMissing},
			},
		},
	}
	out := RenderCommunityToolResult([]communitytool.Result{result}, assertErr("validation failed"))
	for _, want := range []string{"Community tool setup failed", "validation failed", "CodeGraph: CLI missing", "OpenCode: missing"} {
		if !strings.Contains(out, want) {
			t.Fatalf("result output missing %q; output:\n%s", want, out)
		}
	}
}

type assertErr string

func (e assertErr) Error() string { return string(e) }
