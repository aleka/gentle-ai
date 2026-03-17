package filemerge

import (
	"encoding/json"
	"testing"
)

func TestMergeJSONObjectsRecursively(t *testing.T) {
	base := []byte(`{"plugins":["a"],"settings":{"theme":"default","flags":{"x":true}}}`)
	overlay := []byte(`{"settings":{"theme":"gentleman","flags":{"y":true}},"extra":1}`)

	merged, err := MergeJSONObjects(base, overlay)
	if err != nil {
		t.Fatalf("MergeJSONObjects() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(merged, &got); err != nil {
		t.Fatalf("Unmarshal merged json error = %v", err)
	}

	settings := got["settings"].(map[string]any)
	flags := settings["flags"].(map[string]any)

	if settings["theme"] != "gentleman" {
		t.Fatalf("theme = %v", settings["theme"])
	}

	if flags["x"] != true || flags["y"] != true {
		t.Fatalf("flags = %#v", flags)
	}

	plugins := got["plugins"].([]any)
	if len(plugins) != 1 || plugins[0] != "a" {
		t.Fatalf("plugins = %#v", plugins)
	}
}

func TestMergeJSONObjectsSupportsJSONCBase(t *testing.T) {
	base := []byte(`{
	  // VS Code-style comments and trailing commas
	  "editor.fontSize": 14,
	  "files.exclude": {
	    "**/.git": true,
	  },
	}`)
	overlay := []byte(`{"chat.tools.autoApprove": true}`)

	merged, err := MergeJSONObjects(base, overlay)
	if err != nil {
		t.Fatalf("MergeJSONObjects() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(merged, &got); err != nil {
		t.Fatalf("Unmarshal merged json error = %v", err)
	}

	autoApprove, ok := got["chat.tools.autoApprove"].(bool)
	if !ok || !autoApprove {
		t.Fatalf("chat.tools.autoApprove = %#v", got["chat.tools.autoApprove"])
	}

	if got["editor.fontSize"] != float64(14) {
		t.Fatalf("editor.fontSize = %v", got["editor.fontSize"])
	}
}

func TestMergeJSONObjectsMalformedBaseReturnsOverlayOnly(t *testing.T) {
	// Real user machines (e.g. Windows) may have a malformed ~/.cursor/mcp.json.
	// The installer should recover by treating the broken base as {} and continuing.
	tests := []struct {
		name    string
		base    []byte
		overlay []byte
		wantKey string
	}{
		{
			name:    "base starting with letter",
			base:    []byte(`allow: all`),
			overlay: []byte(`{"mcpServers": {"context7": {"type": "remote"}}}`),
			wantKey: "mcpServers",
		},
		{
			name:    "unclosed json object",
			base:    []byte(`{"ok": true`),
			overlay: []byte(`{"chat.tools.autoApprove": true}`),
			wantKey: "chat.tools.autoApprove",
		},
		{
			name:    "arbitrary text",
			base:    []byte(`a`),
			overlay: []byte(`{"servers": {"engram": {"command": "engram"}}}`),
			wantKey: "servers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged, err := MergeJSONObjects(tt.base, tt.overlay)
			if err != nil {
				t.Fatalf("MergeJSONObjects() error = %v; want nil (malformed base should be treated as {})", err)
			}

			var got map[string]any
			if err := json.Unmarshal(merged, &got); err != nil {
				t.Fatalf("merged result is not valid JSON: %v", err)
			}

			if _, ok := got[tt.wantKey]; !ok {
				t.Fatalf("merged result missing key %q from overlay; got keys: %v", tt.wantKey, got)
			}
		})
	}
}
