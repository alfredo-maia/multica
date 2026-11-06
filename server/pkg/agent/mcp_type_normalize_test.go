package agent

import (
	"encoding/json"
	"testing"
)

func TestEnsureMcpServerTypeFields(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want map[string]bool // server name -> expect "type" present after
	}{
		{
			name: "adds type http when url present and type missing",
			in:   `{"mcpServers":{"sankhya-ajuda":{"url":"https://x/","headers":{"Authorization":"Bearer t"}}}}`,
			want: map[string]bool{"sankhya-ajuda": true},
		},
		{
			name: "leaves existing type untouched",
			in:   `{"mcpServers":{"foo":{"type":"sse","url":"https://x/"}}}`,
			want: map[string]bool{"foo": true},
		},
		{
			name: "leaves stdio/command servers untouched (no type added)",
			in:   `{"mcpServers":{"bar":{"command":"npx","args":["x"]}}}`,
			want: map[string]bool{"bar": false},
		},
		{
			name: "no mcpServers key: passthrough",
			in:   `{"foo":"bar"}`,
			want: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := ensureMcpServerTypeFields(json.RawMessage(c.in))
			var doc map[string]json.RawMessage
			if err := json.Unmarshal(out, &doc); err != nil {
				t.Fatalf("output not valid json: %v (%s)", err, out)
			}
			serversRaw, ok := doc["mcpServers"]
			if !ok {
				if c.want != nil {
					t.Fatalf("expected mcpServers key in output")
				}
				return
			}
			var servers map[string]map[string]any
			if err := json.Unmarshal(serversRaw, &servers); err != nil {
				t.Fatalf("mcpServers not valid json: %v", err)
			}
			for name, wantType := range c.want {
				_, hasType := servers[name]["type"]
				if hasType != wantType {
					t.Errorf("server %s: type present=%v, want %v (entry=%v)", name, hasType, wantType, servers[name])
				}
			}
		})
	}
}

func TestEnsureMcpServerTypeFieldsPreservesExistingTypeValue(t *testing.T) {
	in := json.RawMessage(`{"mcpServers":{"foo":{"type":"sse","url":"https://x/"}}}`)
	out := ensureMcpServerTypeFields(in)
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	var servers map[string]map[string]any
	if err := json.Unmarshal(doc["mcpServers"], &servers); err != nil {
		t.Fatalf("invalid mcpServers: %v", err)
	}
	if servers["foo"]["type"] != "sse" {
		t.Errorf("type was overwritten: got %v, want sse", servers["foo"]["type"])
	}
}
