package server

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const defaultDocsURL = "http://mythic_documentation:8090"

// docsBaseURL returns the Mythic documentation container base URL,
// configurable via MYTHIC_DOCS_URL env var.
func docsBaseURL() string {
	if u := os.Getenv("MYTHIC_DOCS_URL"); u != "" {
		return strings.TrimRight(u, "/")
	}
	return defaultDocsURL
}

// docsHTTPClient is a shared HTTP client for fetching documentation.
var docsHTTPClient = &http.Client{
	Timeout: 15 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

// docIndexEntry represents a single entry from the Hugo search index.
type docIndexEntry struct {
	URI         string   `json:"uri"`
	Title       string   `json:"title"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
}

// docIndex caches the fetched documentation index.
type docIndex struct {
	mu      sync.Mutex
	entries []docIndexEntry
	fetched time.Time
	ttl     time.Duration
}

var cachedDocIndex = &docIndex{ttl: 5 * time.Minute}

// fetch retrieves the docs index from the documentation container, with caching.
func (di *docIndex) fetch() ([]docIndexEntry, error) {
	di.mu.Lock()
	defer di.mu.Unlock()

	if di.entries != nil && time.Since(di.fetched) < di.ttl {
		return di.entries, nil
	}

	url := docsBaseURL() + "/docs/index.json"
	resp, err := docsHTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to reach documentation server at %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("documentation server returned %d from %s", resp.StatusCode, url)
	}

	var entries []docIndexEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("failed to parse documentation index: %w", err)
	}

	di.entries = entries
	di.fetched = time.Now()
	return entries, nil
}

// uriToPath converts a Hugo URI like "//host:port/docs/agents/poseidon/"
// to a clean path like "agents/poseidon".
func uriToPath(uri string) string {
	// Strip scheme-relative prefix and host
	if idx := strings.Index(uri, "/docs/"); idx >= 0 {
		uri = uri[idx+len("/docs/"):]
	}
	return strings.Trim(uri, "/")
}

// docTreeNode is used to build a hierarchical tree from flat paths.
type docTreeNode struct {
	Name     string         `json:"name"`
	Path     string         `json:"path"`
	Title    string         `json:"title,omitempty"`
	HasDoc   bool           `json:"has_doc,omitempty"`
	Children []*docTreeNode `json:"children,omitempty"`
}

// registerDocumentationTools registers documentation browsing tools.
func (s *Server) registerDocumentationTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_list_documentation",
		Description: "List available documentation for installed Mythic agents, C2 profiles, and wrappers. Returns a tree of documentation pages fetched from the Mythic documentation server. Use this to discover what documentation is available before retrieving specific pages with mythic_get_documentation. Each entry has a path field you can pass to mythic_get_documentation.",
	}, s.handleListDocumentation)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_documentation",
		Description: "Retrieve the text content of a specific documentation page from the Mythic documentation server. Use mythic_list_documentation first to discover available pages. Pass the path value from the listing (e.g. agents/poseidon, c2-profiles/httpx/examples, agents/poseidon/commands/shell). IMPORTANT: Always read C2 profile documentation before creating payloads - profiles may require specific configuration files or parameters not obvious from the API alone.",
	}, s.handleGetDocumentation)
}

type listDocumentationArgs struct{}

type getDocumentationArgs struct {
	Path string `json:"path" jsonschema:"Documentation path from the listing (e.g. agents/poseidon, c2-profiles/httpx/examples, agents/poseidon/commands/shell)"`
}

func (s *Server) handleListDocumentation(ctx context.Context, req *mcp.CallToolRequest, args listDocumentationArgs) (*mcp.CallToolResult, any, error) {
	entries, err := cachedDocIndex.fetch()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to fetch documentation: %v\nSet MYTHIC_DOCS_URL env var if the documentation server is at a different address (default: %s).", err, defaultDocsURL),
				},
			},
		}, nil, nil
	}

	if len(entries) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "No documentation found."},
			},
		}, nil, nil
	}

	// Build a tree from flat URIs
	root := &docTreeNode{Name: "root"}
	for _, e := range entries {
		path := uriToPath(e.URI)
		if path == "" || path == "categories" || path == "tags" {
			continue
		}
		insertIntoTree(root, path, e.Title)
	}

	// Render as categorized tree text
	var sb strings.Builder
	sb.WriteString("Available Mythic Documentation\n")
	sb.WriteString("==============================\n\n")

	// Sort top-level children for consistent output
	sort.Slice(root.Children, func(i, j int) bool {
		return root.Children[i].Name < root.Children[j].Name
	})

	for _, cat := range root.Children {
		sb.WriteString(fmt.Sprintf("## %s\n", cat.Title))
		writeDocTree(&sb, cat.Children, "")
		sb.WriteString("\n")
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, wrapList(root.Children), nil
}

func insertIntoTree(root *docTreeNode, path, title string) {
	parts := strings.Split(path, "/")
	current := root
	for i, part := range parts {
		found := false
		for _, child := range current.Children {
			if child.Name == part {
				current = child
				found = true
				break
			}
		}
		if !found {
			fullPath := strings.Join(parts[:i+1], "/")
			node := &docTreeNode{
				Name:  part,
				Path:  fullPath,
				Title: part,
			}
			current.Children = append(current.Children, node)
			current = node
		}
	}
	// Set the title and mark as documented for the leaf
	current.Title = title
	current.HasDoc = true
}

func writeDocTree(sb *strings.Builder, nodes []*docTreeNode, indent string) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})
	for i, n := range nodes {
		connector := "|- "
		childIndent := indent + "|  "
		if i == len(nodes)-1 {
			connector = "\\- "
			childIndent = indent + "   "
		}
		suffix := ""
		if len(n.Children) > 0 {
			suffix = fmt.Sprintf(" (%d sub-pages)", countLeaves(n))
		}
		displayName := n.Title
		if displayName == "" {
			displayName = n.Name
		}
		sb.WriteString(fmt.Sprintf("%s%s%s  [path: \"%s\"]%s\n", indent, connector, displayName, n.Path, suffix))
		if len(n.Children) > 0 {
			writeDocTree(sb, n.Children, childIndent)
		}
	}
}

func countLeaves(n *docTreeNode) int {
	if len(n.Children) == 0 {
		return 1
	}
	count := 0
	for _, c := range n.Children {
		count += countLeaves(c)
	}
	return count
}

func (s *Server) handleGetDocumentation(ctx context.Context, req *mcp.CallToolRequest, args getDocumentationArgs) (*mcp.CallToolResult, any, error) {
	if args.Path == "" {
		return nil, nil, fmt.Errorf("path is required")
	}

	entries, err := cachedDocIndex.fetch()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch documentation: %w", err)
	}

	// Normalize requested path
	requested := strings.Trim(strings.ToLower(args.Path), "/")
	// Also support the old-style path format with spaces (e.g. "C2 Profiles/http")
	requested = strings.ReplaceAll(requested, " ", "-")

	// Find exact match first, then prefix matches
	var exactMatch *docIndexEntry
	var prefixMatches []docIndexEntry
	for _, e := range entries {
		entryPath := strings.ToLower(uriToPath(e.URI))
		if entryPath == requested {
			exactMatch = &e
			break
		}
		if strings.HasPrefix(entryPath, requested+"/") {
			prefixMatches = append(prefixMatches, e)
		}
	}

	if exactMatch != nil {
		content := exactMatch.Content
		if content == "" {
			content = fmt.Sprintf("# %s\n\n%s", exactMatch.Title, exactMatch.Description)
		}

		// If this is a "directory" node, also list children
		var childList []string
		for _, e := range entries {
			ep := uriToPath(e.URI)
			if strings.HasPrefix(strings.ToLower(ep), requested+"/") {
				// Direct children only (one level deeper)
				remainder := strings.TrimPrefix(strings.ToLower(ep), requested+"/")
				if !strings.Contains(remainder, "/") && remainder != "" {
					childList = append(childList, fmt.Sprintf("- %s  [path: \"%s\"]", e.Title, ep))
				}
			}
		}

		if len(childList) > 0 {
			content += "\n\n---\n## Sub-Pages\n\n" + strings.Join(childList, "\n")
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: content},
			},
		}, exactMatch, nil
	}

	// No exact match — if we have prefix matches, show the "directory" listing
	if len(prefixMatches) > 0 {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("## Documentation under \"%s\"\n\n", args.Path))
		for _, e := range prefixMatches {
			ep := uriToPath(e.URI)
			sb.WriteString(fmt.Sprintf("- %s  [path: \"%s\"]\n", e.Title, ep))
		}
		sb.WriteString(fmt.Sprintf("\nUse mythic_get_documentation with one of the paths above to read the full content."))
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: sb.String()},
			},
		}, wrapList(prefixMatches), nil
	}

	return nil, nil, fmt.Errorf("documentation not found at path: %s — use mythic_list_documentation or mythic_search_documentation to discover valid paths", args.Path)
}
