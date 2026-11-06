package agent

import "encoding/json"

// ensureMcpServerTypeFields fills in a missing "type" on HTTP-style MCP
// server entries (those with a "url" and no "command") before the config is
// handed to the provider CLI.
//
// Root cause: workspace-library MCP servers are stored as
// {"url": "...", "headers": {...}} without an explicit "type". Older Claude
// Code releases inferred the transport from the presence of "url"; current
// releases (>= 2.1.x) require an explicit "type": "http" (or "sse"/"ws") and
// silently DROP any server entry missing it, reporting
// mcp_server_errors: [{"type": "url_missing_type", ...}] with no tools
// loaded and no hard failure — this is the actual cause of MCP HTTP servers
// going missing in fresh Claude Code sessions, distinct from the
// --allowedTools race condition addressed by PR #4907.
//
// This only ADDS a "type" field when one is absent; it never removes or
// overrides an existing "type", and it never touches stdio ("command")
// servers. Safe no-op for every config that is already well-formed.
func ensureMcpServerTypeFields(raw json.RawMessage) json.RawMessage {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		return raw
	}
	serversRaw, ok := top["mcpServers"]
	if !ok {
		return raw
	}
	var servers map[string]json.RawMessage
	if err := json.Unmarshal(serversRaw, &servers); err != nil {
		return raw
	}

	changed := false
	for name, serverRaw := range servers {
		var entry map[string]any
		if err := json.Unmarshal(serverRaw, &entry); err != nil {
			continue
		}
		if _, hasType := entry["type"]; hasType {
			continue
		}
		if _, hasCommand := entry["command"]; hasCommand {
			continue
		}
		urlVal, hasURL := entry["url"]
		if !hasURL {
			continue
		}
		if s, ok := urlVal.(string); !ok || s == "" {
			continue
		}
		entry["type"] = "http"
		servers[name], changed = mustMarshalRaw(entry), true
	}
	if !changed {
		return raw
	}

	top["mcpServers"] = mustMarshalRaw(servers)
	data, err := json.Marshal(top)
	if err != nil {
		return raw
	}
	return data
}
