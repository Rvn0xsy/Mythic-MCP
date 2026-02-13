package server

// wrapList wraps a slice in a JSON-object-compatible map for use as
// MCP structuredContent. The MCP protocol requires structuredContent
// to be a JSON object (not an array), so list responses must be
// wrapped in an envelope like {"items": [...], "count": N}.
func wrapList[T any](items []T) map[string]any {
	return map[string]any{
		"items": items,
		"count": len(items),
	}
}
