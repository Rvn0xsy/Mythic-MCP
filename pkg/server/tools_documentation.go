package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const defaultDocsPath = "/root/mythic/documentation-docker/content"

// docsBasePath returns the documentation content directory,
// configurable via MYTHIC_DOCS_PATH env var.
func docsBasePath() string {
	if p := os.Getenv("MYTHIC_DOCS_PATH"); p != "" {
		return p
	}
	return defaultDocsPath
}

// registerDocumentationTools registers documentation browsing tools.
func (s *Server) registerDocumentationTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_list_documentation",
		Description: "List available documentation for installed Mythic agents, C2 profiles, and wrappers. Returns a tree of documentation pages. Use this to discover what documentation is available before retrieving specific pages with mythic_get_documentation. Each entry has a path field you can pass to mythic_get_documentation.",
	}, s.handleListDocumentation)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_documentation",
		Description: "Retrieve the full markdown content of a specific documentation page. Use mythic_list_documentation first to discover available pages. Pass the path value from the listing (e.g. Agents/poseidon, C2 Profiles/httpx/examples, Agents/poseidon/commands/shell). IMPORTANT: Always read C2 profile documentation before creating payloads - profiles may require specific configuration files or parameters not obvious from the API alone.",
	}, s.handleGetDocumentation)
}

type listDocumentationArgs struct{}

type getDocumentationArgs struct {
	Path string `json:"path" jsonschema:"Documentation path from the listing (e.g. Agents/poseidon or C2 Profiles/httpx/examples or Agents/poseidon/commands/shell)"`
}

// docEntry represents a single documentation node in the tree.
type docEntry struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	Children []docEntry `json:"children,omitempty"`
}

// buildDocTree walks a directory and builds a tree of doc entries.
func buildDocTree(basePath, relPath string) []docEntry {
	absPath := filepath.Join(basePath, relPath)
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil
	}

	var result []docEntry
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		childRel := filepath.Join(relPath, e.Name())
		children := buildDocTree(basePath, childRel)

		hasDoc := false
		dirEntries, _ := os.ReadDir(filepath.Join(basePath, childRel))
		for _, de := range dirEntries {
			if !de.IsDir() && strings.HasSuffix(de.Name(), ".md") {
				hasDoc = true
				break
			}
		}

		if hasDoc || len(children) > 0 {
			entry := docEntry{
				Name:     e.Name(),
				Path:     childRel,
				Children: children,
			}
			result = append(result, entry)
		}
	}
	return result
}

func (s *Server) handleListDocumentation(ctx context.Context, req *mcp.CallToolRequest, args listDocumentationArgs) (*mcp.CallToolResult, any, error) {
	base := docsBasePath()
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Documentation directory not found at %s. Set MYTHIC_DOCS_PATH env var if docs are at a different location.", base),
				},
			},
		}, nil, nil
	}

	type category struct {
		Name    string     `json:"category"`
		Entries []docEntry `json:"entries"`
	}

	categories := []string{"Agents", "C2 Profiles", "Wrappers"}
	var result []category
	for _, cat := range categories {
		catPath := filepath.Join(base, cat)
		if _, err := os.Stat(catPath); os.IsNotExist(err) {
			continue
		}
		entries := buildDocTree(base, cat)
		if len(entries) > 0 {
			result = append(result, category{Name: cat, Entries: entries})
		}
	}

	if len(result) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "No documentation found."},
			},
		}, nil, nil
	}

	var sb strings.Builder
	sb.WriteString("Available Mythic Documentation\n")
	sb.WriteString("==============================\n\n")
	for _, cat := range result {
		sb.WriteString(fmt.Sprintf("## %s\n", cat.Name))
		writeTree(&sb, cat.Entries, "")
		sb.WriteString("\n")
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, result, nil
}

func writeTree(sb *strings.Builder, entries []docEntry, indent string) {
	for i, e := range entries {
		connector := "|- "
		childIndent := indent + "|  "
		if i == len(entries)-1 {
			connector = "\\- "
			childIndent = indent + "   "
		}
		pages := listPages(filepath.Join(docsBasePath(), e.Path))
		pageSuffix := ""
		if len(pages) > 0 {
			pageSuffix = fmt.Sprintf(" (%d pages)", len(pages))
		}
		sb.WriteString(fmt.Sprintf("%s%s%s%s\n", indent, connector, e.Name, pageSuffix))
		if len(e.Children) > 0 {
			writeTree(sb, e.Children, childIndent)
		}
	}
}

func listPages(dirPath string) []string {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil
	}
	var pages []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		if e.Name() == "_index.md" {
			continue
		}
		pages = append(pages, strings.TrimSuffix(e.Name(), ".md"))
	}
	return pages
}

func (s *Server) handleGetDocumentation(ctx context.Context, req *mcp.CallToolRequest, args getDocumentationArgs) (*mcp.CallToolResult, any, error) {
	if args.Path == "" {
		return nil, nil, fmt.Errorf("path is required")
	}

	base := docsBasePath()
	cleaned := filepath.Clean(args.Path)
	if strings.Contains(cleaned, "..") {
		return nil, nil, fmt.Errorf("invalid path")
	}

	target := filepath.Join(base, cleaned)

	info, err := os.Stat(target)
	if err != nil && os.IsNotExist(err) {
		// try appending .md
		mdPath := target + ".md"
		if content, readErr := os.ReadFile(mdPath); readErr == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(content)},
				},
			}, nil, nil
		}
		return nil, nil, fmt.Errorf("documentation not found at path: %s", cleaned)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to access documentation: %w", err)
	}

	var sb strings.Builder

	if info.IsDir() {
		indexPath := filepath.Join(target, "_index.md")
		if indexContent, readErr := os.ReadFile(indexPath); readErr == nil {
			sb.WriteString(string(indexContent))
			sb.WriteString("\n\n")
		}

		dirEntries, _ := os.ReadDir(target)
		var subPages []string
		var subDirs []string
		for _, e := range dirEntries {
			if e.IsDir() {
				subDirs = append(subDirs, e.Name())
			} else if strings.HasSuffix(e.Name(), ".md") && e.Name() != "_index.md" {
				subPages = append(subPages, strings.TrimSuffix(e.Name(), ".md"))
			}
		}

		if len(subPages) > 0 || len(subDirs) > 0 {
			sb.WriteString("---\n## Available Sub-Pages\n\n")
			for _, p := range subPages {
				sb.WriteString(fmt.Sprintf("- %s (path: \"%s/%s\")\n", p, cleaned, p))
			}
			for _, d := range subDirs {
				sb.WriteString(fmt.Sprintf("- %s/ (path: \"%s/%s\")\n", d, cleaned, d))
			}
		}
	} else {
		content, readErr := os.ReadFile(target)
		if readErr != nil {
			return nil, nil, fmt.Errorf("failed to read documentation: %w", readErr)
		}
		sb.WriteString(string(content))
	}

	if sb.Len() == 0 {
		return nil, nil, fmt.Errorf("documentation not found at path: %s", cleaned)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}
