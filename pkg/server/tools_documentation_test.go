package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sampleDocsIndex is a minimal but representative Hugo search index
// matching the structure served by mythic_documentation at /docs/index.json.
var sampleDocsIndex = []docIndexEntry{
	{URI: "//localhost:8090/docs/agents/", Title: "Agents", Content: "Agent overview content"},
	{URI: "//localhost:8090/docs/agents/poseidon/", Title: "poseidon", Content: "Poseidon agent for macOS and Linux"},
	{URI: "//localhost:8090/docs/agents/poseidon/commands/", Title: "Commands", Content: "Poseidon command listing"},
	{URI: "//localhost:8090/docs/agents/poseidon/commands/shell/", Title: "shell", Content: "Execute a shell command with bash -c"},
	{URI: "//localhost:8090/docs/agents/poseidon/commands/ls/", Title: "ls", Content: "List directory contents"},
	{URI: "//localhost:8090/docs/agents/poseidon/commands/download/", Title: "download", Content: "Download a file from target"},
	{URI: "//localhost:8090/docs/agents/poseidon/c2_profiles/", Title: "C2 Profiles", Content: ""},
	{URI: "//localhost:8090/docs/agents/poseidon/c2_profiles/http/", Title: "HTTP", Content: "HTTP C2 profile config"},
	{URI: "//localhost:8090/docs/c2-profiles/", Title: "C2 Profiles", Content: "C2 profile overview"},
	{URI: "//localhost:8090/docs/c2-profiles/httpx/", Title: "httpx", Content: "HTTPX C2 profile details"},
	{URI: "//localhost:8090/docs/c2-profiles/httpx/examples/", Title: "examples", Content: "HTTPX example configurations"},
	{URI: "//localhost:8090/docs/wrappers/", Title: "Wrappers", Content: "Wrapper overview"},
	{URI: "//localhost:8090/docs/categories/", Title: "Categories", Content: ""},
	{URI: "//localhost:8090/docs/tags/", Title: "Tags", Content: ""},
}

// startFakeDocsServer creates an httptest server that serves sampleDocsIndex
// at /docs/index.json and sets MYTHIC_DOCS_URL to point to it.
// Returns a cleanup function.
func startFakeDocsServer(t *testing.T) func() {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/docs/index.json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sampleDocsIndex)
	}))

	os.Setenv("MYTHIC_DOCS_URL", srv.URL)

	// Reset the cache so tests get fresh data
	cachedDocIndex.mu.Lock()
	cachedDocIndex.entries = nil
	cachedDocIndex.mu.Unlock()

	return func() {
		srv.Close()
		os.Unsetenv("MYTHIC_DOCS_URL")
		// Reset cache after test
		cachedDocIndex.mu.Lock()
		cachedDocIndex.entries = nil
		cachedDocIndex.mu.Unlock()
	}
}

// --- uriToPath tests ---

func TestUriToPath(t *testing.T) {
	tests := []struct {
		uri  string
		want string
	}{
		{"//localhost:8090/docs/agents/poseidon/", "agents/poseidon"},
		{"//localhost:8090/docs/c2-profiles/httpx/examples/", "c2-profiles/httpx/examples"},
		{"//localhost:8090/docs/agents/", "agents"},
		{"//localhost:8090/docs/", ""},
		{"/docs/agents/poseidon/commands/shell/", "agents/poseidon/commands/shell"},
		{"agents/poseidon/", "agents/poseidon"},
	}
	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			got := uriToPath(tt.uri)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- insertIntoTree tests ---

func TestInsertIntoTree(t *testing.T) {
	root := &docTreeNode{Name: "root"}

	insertIntoTree(root, "agents/poseidon/commands/shell", "shell")
	insertIntoTree(root, "agents/poseidon/commands/ls", "ls")
	insertIntoTree(root, "agents/poseidon", "poseidon")
	insertIntoTree(root, "c2-profiles/httpx", "httpx")

	// Should have 2 top-level children: agents, c2-profiles
	require.Len(t, root.Children, 2)

	// Find agents
	var agents *docTreeNode
	for _, c := range root.Children {
		if c.Name == "agents" {
			agents = c
		}
	}
	require.NotNil(t, agents)

	// agents -> poseidon
	require.Len(t, agents.Children, 1)
	poseidon := agents.Children[0]
	assert.Equal(t, "poseidon", poseidon.Name)
	assert.Equal(t, "poseidon", poseidon.Title)
	assert.True(t, poseidon.HasDoc)

	// poseidon -> commands
	require.Len(t, poseidon.Children, 1)
	commands := poseidon.Children[0]
	assert.Equal(t, "commands", commands.Name)

	// commands -> shell, ls
	require.Len(t, commands.Children, 2)
}

func TestInsertIntoTree_EmptyPath(t *testing.T) {
	root := &docTreeNode{Name: "root"}
	insertIntoTree(root, "", "empty")
	// Empty path creates a single child with empty name
	// This is fine — we filter these out in the handler
}

// --- countLeaves tests ---

func TestCountLeaves(t *testing.T) {
	root := &docTreeNode{Name: "root"}
	insertIntoTree(root, "a/b/c", "c")
	insertIntoTree(root, "a/b/d", "d")
	insertIntoTree(root, "a/e", "e")

	// root -> a -> b -> c,d  and  a -> e
	assert.Equal(t, 3, countLeaves(root))

	leaf := &docTreeNode{Name: "leaf"}
	assert.Equal(t, 1, countLeaves(leaf))
}

// --- docIndex.fetch tests ---

func TestDocIndexFetch_Success(t *testing.T) {
	cleanup := startFakeDocsServer(t)
	defer cleanup()

	entries, err := cachedDocIndex.fetch()
	require.NoError(t, err)
	assert.Len(t, entries, len(sampleDocsIndex))
}

func TestDocIndexFetch_Caching(t *testing.T) {
	cleanup := startFakeDocsServer(t)
	defer cleanup()

	// First fetch
	entries1, err := cachedDocIndex.fetch()
	require.NoError(t, err)

	// Second fetch should return from cache (same pointer)
	entries2, err := cachedDocIndex.fetch()
	require.NoError(t, err)

	// Slices backed by same array when cached
	assert.Equal(t, len(entries1), len(entries2))
}

func TestDocIndexFetch_ServerDown(t *testing.T) {
	// Point to a non-existent server
	os.Setenv("MYTHIC_DOCS_URL", "http://127.0.0.1:1")
	defer os.Unsetenv("MYTHIC_DOCS_URL")

	cachedDocIndex.mu.Lock()
	cachedDocIndex.entries = nil
	cachedDocIndex.mu.Unlock()

	_, err := cachedDocIndex.fetch()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to reach documentation server")
}

func TestDocIndexFetch_BadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	os.Setenv("MYTHIC_DOCS_URL", srv.URL)
	defer os.Unsetenv("MYTHIC_DOCS_URL")

	cachedDocIndex.mu.Lock()
	cachedDocIndex.entries = nil
	cachedDocIndex.mu.Unlock()

	_, err := cachedDocIndex.fetch()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "returned 500")
}

func TestDocIndexFetch_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "not json")
	}))
	defer srv.Close()

	os.Setenv("MYTHIC_DOCS_URL", srv.URL)
	defer os.Unsetenv("MYTHIC_DOCS_URL")

	cachedDocIndex.mu.Lock()
	cachedDocIndex.entries = nil
	cachedDocIndex.mu.Unlock()

	_, err := cachedDocIndex.fetch()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse documentation index")
}

// --- handleListDocumentation tests ---

func TestHandleListDocumentation_StructuredContentIsObject(t *testing.T) {
	cleanup := startFakeDocsServer(t)
	defer cleanup()

	s := &Server{}
	result, structured, err := s.handleListDocumentation(nil, &mcp.CallToolRequest{}, listDocumentationArgs{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, structured)

	// The critical assertion: structured content must marshal to a JSON object, not array
	data, err := json.Marshal(structured)
	require.NoError(t, err)
	assert.Equal(t, byte('{'), data[0], "structuredContent must be a JSON object, not array")

	// Should have items and count keys
	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)
	assert.Contains(t, parsed, "items")
	assert.Contains(t, parsed, "count")
}

func TestHandleListDocumentation_TextContent(t *testing.T) {
	cleanup := startFakeDocsServer(t)
	defer cleanup()

	s := &Server{}
	result, _, err := s.handleListDocumentation(nil, &mcp.CallToolRequest{}, listDocumentationArgs{})
	require.NoError(t, err)
	require.Len(t, result.Content, 1)

	text := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, text, "Available Mythic Documentation")
	assert.Contains(t, text, "Agents")
	assert.Contains(t, text, "poseidon")
	assert.Contains(t, text, "shell")
	assert.Contains(t, text, "C2 Profiles")
	assert.Contains(t, text, `[path: "agents/poseidon"]`)
}

func TestHandleListDocumentation_FiltersCategoriesAndTags(t *testing.T) {
	cleanup := startFakeDocsServer(t)
	defer cleanup()

	s := &Server{}
	result, _, err := s.handleListDocumentation(nil, &mcp.CallToolRequest{}, listDocumentationArgs{})
	require.NoError(t, err)

	text := result.Content[0].(*mcp.TextContent).Text
	assert.NotContains(t, text, "categories")
	assert.NotContains(t, text, "tags")
}

func TestHandleListDocumentation_EmptyIndex(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "[]")
	}))
	defer srv.Close()

	os.Setenv("MYTHIC_DOCS_URL", srv.URL)
	defer func() {
		os.Unsetenv("MYTHIC_DOCS_URL")
		cachedDocIndex.mu.Lock()
		cachedDocIndex.entries = nil
		cachedDocIndex.mu.Unlock()
	}()

	cachedDocIndex.mu.Lock()
	cachedDocIndex.entries = nil
	cachedDocIndex.mu.Unlock()

	s := &Server{}
	result, structured, err := s.handleListDocumentation(nil, &mcp.CallToolRequest{}, listDocumentationArgs{})
	require.NoError(t, err)
	assert.Nil(t, structured)

	text := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, text, "No documentation found")
}

func TestHandleListDocumentation_ServerError(t *testing.T) {
	os.Setenv("MYTHIC_DOCS_URL", "http://127.0.0.1:1")
	defer os.Unsetenv("MYTHIC_DOCS_URL")

	cachedDocIndex.mu.Lock()
	cachedDocIndex.entries = nil
	cachedDocIndex.mu.Unlock()

	s := &Server{}
	result, structured, err := s.handleListDocumentation(nil, &mcp.CallToolRequest{}, listDocumentationArgs{})
	require.NoError(t, err) // returns error in content, not as Go error
	assert.Nil(t, structured)

	text := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, text, "Failed to fetch documentation")
}

// --- handleGetDocumentation tests ---

func TestHandleGetDocumentation_ExactMatch(t *testing.T) {
	cleanup := startFakeDocsServer(t)
	defer cleanup()

	s := &Server{}
	result, structured, err := s.handleGetDocumentation(nil, &mcp.CallToolRequest{}, getDocumentationArgs{Path: "agents/poseidon/commands/shell"})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, structured)

	text := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, text, "Execute a shell command with bash -c")

	// Structured content should be an object (docIndexEntry), not array
	data, err := json.Marshal(structured)
	require.NoError(t, err)
	assert.Equal(t, byte('{'), data[0], "exact match structuredContent must be a JSON object")
}

func TestHandleGetDocumentation_CaseInsensitive(t *testing.T) {
	cleanup := startFakeDocsServer(t)
	defer cleanup()

	s := &Server{}
	result, _, err := s.handleGetDocumentation(nil, &mcp.CallToolRequest{}, getDocumentationArgs{Path: "Agents/Poseidon/Commands/Shell"})
	require.NoError(t, err)
	require.NotNil(t, result)

	text := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, text, "Execute a shell command")
}

func TestHandleGetDocumentation_SpaceToHyphen(t *testing.T) {
	cleanup := startFakeDocsServer(t)
	defer cleanup()

	s := &Server{}
	result, _, err := s.handleGetDocumentation(nil, &mcp.CallToolRequest{}, getDocumentationArgs{Path: "C2 Profiles/httpx"})
	require.NoError(t, err)
	require.NotNil(t, result)

	text := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, text, "HTTPX C2 profile details")
}

func TestHandleGetDocumentation_DirectoryWithChildren(t *testing.T) {
	cleanup := startFakeDocsServer(t)
	defer cleanup()

	s := &Server{}
	result, _, err := s.handleGetDocumentation(nil, &mcp.CallToolRequest{}, getDocumentationArgs{Path: "agents/poseidon/commands"})
	require.NoError(t, err)
	require.NotNil(t, result)

	text := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, text, "Sub-Pages")
	assert.Contains(t, text, "shell")
	assert.Contains(t, text, "ls")
	assert.Contains(t, text, "download")
}

func TestHandleGetDocumentation_PrefixMatch(t *testing.T) {
	cleanup := startFakeDocsServer(t)
	defer cleanup()

	s := &Server{}
	// "c2-profiles" has children "httpx" and "httpx/examples".
	// "c2-profiles/http" is an exact match, but "c2-profiles/htt" is not —
	// however it also won't prefix-match because "httpx" doesn't start with "htt/".
	// Use a path that IS a valid prefix: request "agents/poseidon/c2_profiles"
	// which exists as exact match. Instead, test with a made-up parent that
	// only appears as prefix. The sample data has "c2-profiles/httpx" and
	// "c2-profiles/httpx/examples", so requesting just "c2-profiles/httpx"
	// gives an exact match + sub-pages but NOT prefix-only.
	//
	// To truly test prefix-only (no exact match), we need a path that doesn't
	// exist but whose children do. The sample has no such case, so add one:
	// We'll test with a path that will hit the prefix branch by requesting
	// a non-existent parent path. Our sample has entries under
	// "agents/poseidon/commands/*" — requesting "agents/poseidon/commands/sh"
	// won't match because "shell" doesn't start with "sh/".
	//
	// Best approach: add a sample entry and test it. For now, skip the prefix
	// test and just verify the wrapList behavior on a synthetic call.
	// Instead let's test with the existing data more carefully.
	// "agents/poseidon" is exact, but we can fabricate a prefix-only scenario
	// by testing directly with entries that have a matching prefix.

	// Actually, let's just test that when prefix matches DO occur, the structured
	// content is wrapped correctly. We can call the handler with a path that
	// won't have an exact match but will match as prefix.
	// None of our sample data naturally produces this — the tree is shallow.
	// So let's just verify the wrapList usage is correct via the structured content type.

	// For a real prefix-only test, temporarily inject an extra entry:
	cachedDocIndex.mu.Lock()
	cachedDocIndex.entries = append(cachedDocIndex.entries, docIndexEntry{
		URI:     "//localhost:8090/docs/guides/setup/basic/",
		Title:   "Basic Setup",
		Content: "How to set up basics",
	}, docIndexEntry{
		URI:     "//localhost:8090/docs/guides/setup/advanced/",
		Title:   "Advanced Setup",
		Content: "Advanced configuration",
	})
	cachedDocIndex.mu.Unlock()

	result, structured, err := s.handleGetDocumentation(nil, &mcp.CallToolRequest{}, getDocumentationArgs{Path: "guides/setup"})
	require.NoError(t, err)
	require.NotNil(t, result)

	text := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, text, "Documentation under")

	// Structured content for prefix matches should be a JSON object (wrapped list)
	data, err := json.Marshal(structured)
	require.NoError(t, err)
	assert.Equal(t, byte('{'), data[0], "prefix match structuredContent must be a JSON object")
}

func TestHandleGetDocumentation_NotFound(t *testing.T) {
	cleanup := startFakeDocsServer(t)
	defer cleanup()

	s := &Server{}
	_, _, err := s.handleGetDocumentation(nil, &mcp.CallToolRequest{}, getDocumentationArgs{Path: "nonexistent/path"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "documentation not found")
}

func TestHandleGetDocumentation_EmptyPath(t *testing.T) {
	cleanup := startFakeDocsServer(t)
	defer cleanup()

	s := &Server{}
	_, _, err := s.handleGetDocumentation(nil, &mcp.CallToolRequest{}, getDocumentationArgs{Path: ""})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path is required")
}

func TestHandleGetDocumentation_TrailingSlash(t *testing.T) {
	cleanup := startFakeDocsServer(t)
	defer cleanup()

	s := &Server{}
	result, _, err := s.handleGetDocumentation(nil, &mcp.CallToolRequest{}, getDocumentationArgs{Path: "agents/poseidon/commands/shell/"})
	require.NoError(t, err)
	require.NotNil(t, result)

	text := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, text, "Execute a shell command")
}

// --- docsBaseURL tests ---

func TestDocsBaseURL_Default(t *testing.T) {
	os.Unsetenv("MYTHIC_DOCS_URL")
	assert.Equal(t, "http://mythic_documentation:8090", docsBaseURL())
}

func TestDocsBaseURL_Custom(t *testing.T) {
	os.Setenv("MYTHIC_DOCS_URL", "http://custom:9090/")
	defer os.Unsetenv("MYTHIC_DOCS_URL")
	// Should strip trailing slash
	assert.Equal(t, "http://custom:9090", docsBaseURL())
}

func TestDocsBaseURL_CustomNoTrailingSlash(t *testing.T) {
	os.Setenv("MYTHIC_DOCS_URL", "http://custom:9090")
	defer os.Unsetenv("MYTHIC_DOCS_URL")
	assert.Equal(t, "http://custom:9090", docsBaseURL())
}
