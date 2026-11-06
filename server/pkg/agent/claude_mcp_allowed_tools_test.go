package agent

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestClaudeMcpAllowedTools(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		want    []string
		wantErr bool
	}{
		{name: "empty", raw: ``, want: nil},
		{name: "null", raw: `null`, want: nil},
		{name: "no servers", raw: `{}`, want: nil},
		{name: "empty servers", raw: `{"mcpServers":{}}`, want: nil},
		{
			name: "sorted wildcard per server",
			raw:  `{"mcpServers":{"sankhya_mcp":{"type":"http","url":"https://x"},"atlassian":{"command":"npx"}}}`,
			want: []string{"mcp__atlassian__*", "mcp__sankhya_mcp__*"},
		},
		{name: "servers not object", raw: `{"mcpServers":[]}`, wantErr: true},
		{name: "server entry not object", raw: `{"mcpServers":{"a":"x"}}`, wantErr: true},
		{name: "comma in name", raw: `{"mcpServers":{"a,b":{"url":"https://x"}}}`, wantErr: true},
		{name: "space in name", raw: `{"mcpServers":{"a b":{"url":"https://x"}}}`, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := claudeMcpAllowedTools(json.RawMessage(tc.raw))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestClaudeMcpAllowedToolsDoesNotLeakSecrets(t *testing.T) {
	raw := `{"mcpServers":{"gw":{"type":"http","url":"https://secret.example/mcp","headers":{"Authorization":"Bearer tok"}}}}`
	got, err := claudeMcpAllowedTools(json.RawMessage(raw))
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range got {
		if tool != "mcp__gw__*" {
			t.Fatalf("unexpected token %q", tool)
		}
	}
}
