// Command gen-schema-docs parses the Mythic-MCP tools_*.go source files
// and generates Markdown documentation pages for each tool category.
//
// Usage:
//
//	go run ./tools/gen-schema-docs            # writes to site/content/tools/
//	go run ./tools/gen-schema-docs -out /tmp  # custom output dir
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
)

// ---------------------------------------------------------------------------
// Domain types
// ---------------------------------------------------------------------------

// Tool represents a single MCP tool extracted from source.
type Tool struct {
	Name        string
	Description string
	Params      []Param
}

// Param is one field from the args struct.
type Param struct {
	Name        string
	Type        string
	JSONName    string
	Description string // from jsonschema tag
	Required    bool
}

// Category groups tools under a heading.
type Category struct {
	Slug        string // filename without extension
	Title       string
	Description string
	Tools       []Tool
}

// ---------------------------------------------------------------------------
// File → category mapping
// ---------------------------------------------------------------------------

var fileMeta = map[string]struct{ title, desc string }{
	"tools_auth":                  {"Authentication", "Login, logout, API token lifecycle, session management."},
	"tools_operations":            {"Operations", "Operation (campaign) CRUD, event logging, global settings."},
	"tools_operators":             {"Operators", "User account management, preferences, secrets, invite links."},
	"tools_callbacks":             {"Callbacks", "Active agent sessions — list, update, P2P edges, tokens."},
	"tools_tasks":                 {"Tasks & Responses", "Issue commands to agents, read output, wait for completion, OPSEC bypass."},
	"tools_payloads":              {"Payloads", "Build, download, manage, and inspect agent payload binaries."},
	"tools_payload_discovery":     {"Payload Discovery", "Introspect build parameters, C2 profile parameters, and available commands for each payload type."},
	"tools_c2profiles":            {"C2 Profiles", "C2 profile lifecycle — start/stop listeners, IOCs, sample messages, configuration."},
	"tools_commands":              {"Commands", "Command schema and parameter introspection."},
	"tools_files":                 {"Files", "Upload, download, preview, and bulk-export files."},
	"tools_credentials_artifacts": {"Credentials & Artifacts", "Credential store and artifact (IOC / forensic evidence) tracking."},
	"tools_attack":                {"MITRE ATT&CK", "Technique lookup, task/command/operation mapping."},
	"tools_hosts":                 {"Hosts", "Host inventory, network topology mapping."},
	"tools_processes":             {"Processes", "Process enumeration and tree views."},
	"tools_screenshots":           {"Screenshots", "Screenshot capture timeline, thumbnails, downloads."},
	"tools_keylogs":               {"Keylogs", "Keylogger data retrieval by operation or callback."},
	"tools_tags":                  {"Tags", "Flexible tagging system for any Mythic object."},
	"tools_documentation":         {"Documentation", "Browse installed agent and C2 profile documentation."},
}

// navSlug maps file stems to the nav slug used in mkdocs.yml.
var navSlug = map[string]string{
	"tools_auth":                  "authentication",
	"tools_operations":            "operations",
	"tools_operators":             "operators",
	"tools_callbacks":             "callbacks",
	"tools_tasks":                 "tasks",
	"tools_payloads":              "payloads",
	"tools_payload_discovery":     "payload-discovery",
	"tools_c2profiles":            "c2-profiles",
	"tools_commands":              "commands",
	"tools_files":                 "files",
	"tools_credentials_artifacts": "credentials",
	"tools_attack":                "mitre-attack",
	"tools_hosts":                 "hosts",
	"tools_processes":             "processes",
	"tools_screenshots":           "screenshots",
	"tools_keylogs":               "keylogs",
	"tools_tags":                  "tags",
	"tools_documentation":         "documentation",
}

// ---------------------------------------------------------------------------
// AST extraction
// ---------------------------------------------------------------------------

// extractTools parses a single Go file and returns tools + arg struct map.
func extractTools(path string) ([]Tool, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	// Pass 1: collect all args structs (name → fields).
	argStructs := map[string][]Param{}
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, spec := range gd.Specs {
			ts := spec.(*ast.TypeSpec)
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}
			name := ts.Name.Name
			if !strings.HasSuffix(name, "Args") && !strings.HasSuffix(name, "args") {
				continue
			}
			var params []Param
			for _, field := range st.Fields.List {
				if len(field.Names) == 0 {
					continue
				}
				p := Param{
					Name:     field.Names[0].Name,
					Type:     typeString(field.Type),
					Required: true,
				}
				if field.Tag != nil {
					p.JSONName = extractTag(field.Tag.Value, "json")
					p.Description = extractTag(field.Tag.Value, "jsonschema")
					if strings.Contains(field.Tag.Value, "omitempty") {
						p.Required = false
					}
				}
				if p.JSONName == "" || p.JSONName == "-" {
					p.JSONName = strings.ToLower(p.Name)
				}
				params = append(params, p)
			}
			argStructs[name] = params
		}
	}

	// Pass 2: find mcp.AddTool calls → extract Name, Description, handler name → match struct.
	var tools []Tool
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "AddTool" {
			return true
		}
		if len(call.Args) < 3 {
			return true
		}

		// Second arg is &mcp.Tool{...}
		toolLit := unaryAddr(call.Args[1])
		if toolLit == nil {
			return true
		}
		t := Tool{}
		for _, elt := range toolLit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			key := fmt.Sprintf("%s", kv.Key)
			switch key {
			case "Name":
				t.Name = stringVal(kv.Value)
			case "Description":
				t.Description = stringVal(kv.Value)
			}
		}

		// Third arg is the handler (s.handleXxx) — derive struct name.
		handlerName := selectorName(call.Args[2])
		if handlerName != "" {
			// Convention: handleFoo → fooArgs  (lowercase first letter)
			structName := strings.TrimPrefix(handlerName, "handle")
			structName = strings.ToLower(structName[:1]) + structName[1:] + "Args"
			if params, ok := argStructs[structName]; ok {
				t.Params = params
			}
			// Also try exact case match
			structName2 := strings.TrimPrefix(handlerName, "handle") + "Args"
			if params, ok := argStructs[structName2]; ok && len(t.Params) == 0 {
				t.Params = params
			}
		}

		if t.Name != "" {
			tools = append(tools, t)
		}
		return true
	})
	return tools, nil
}

func unaryAddr(expr ast.Expr) *ast.CompositeLit {
	ue, ok := expr.(*ast.UnaryExpr)
	if !ok {
		return nil
	}
	cl, ok := ue.X.(*ast.CompositeLit)
	if !ok {
		return nil
	}
	return cl
}

func selectorName(expr ast.Expr) string {
	sel, ok := expr.(*ast.SelectorExpr)
	if ok {
		return sel.Sel.Name
	}
	return ""
}

func stringVal(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.BasicLit:
		return strings.Trim(v.Value, `"`)
	case *ast.BinaryExpr:
		// string concatenation with +
		return stringVal(v.X) + stringVal(v.Y)
	}
	return ""
}

var tagRe = regexp.MustCompile(`(\w+):"([^"]*)"`)

func extractTag(raw, key string) string {
	for _, m := range tagRe.FindAllStringSubmatch(raw, -1) {
		if m[1] == key {
			parts := strings.Split(m[2], ",")
			return parts[0]
		}
	}
	return ""
}

func typeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.ArrayType:
		return "[]" + typeString(t.Elt)
	case *ast.MapType:
		return "map[" + typeString(t.Key) + "]" + typeString(t.Value)
	case *ast.SelectorExpr:
		return typeString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + typeString(t.X)
	case *ast.InterfaceType:
		return "any"
	}
	return "unknown"
}

// ---------------------------------------------------------------------------
// Markdown generation
// ---------------------------------------------------------------------------

var mdTemplate = template.Must(template.New("tool-page").Funcs(template.FuncMap{
	"requiredBadge": func(r bool) string {
		if r {
			return ":material-check-bold:{ title=\"Required\" } **required**"
		}
		return "_optional_"
	},
}).Parse(`# {{ .Title }}

{{ .Description }}

!!! info "Auto-generated"
    This page is **generated from source code** by the schema extractor.
    Do not edit by hand — changes will be overwritten.

---

{{ range .Tools }}
## ` + "`{{ .Name }}`" + `

{{ .Description }}

{{ if .Params -}}
### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
{{ range .Params -}}
| ` + "`{{ .JSONName }}`" + ` | ` + "`{{ .Type }}`" + ` | {{ requiredBadge .Required }} | {{ .Description }} |
{{ end }}
{{ else -}}
_No parameters._
{{ end }}
---

{{ end }}
`))

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	outDir := "site/content/tools"
	if len(os.Args) > 2 && os.Args[1] == "-out" {
		outDir = os.Args[2]
	}

	srcDir := "pkg/server"
	matches, err := filepath.Glob(filepath.Join(srcDir, "tools_*.go"))
	if err != nil {
		fatal(err)
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fatal(err)
	}

	var allCategories []Category
	totalTools := 0

	for _, path := range matches {
		stem := strings.TrimSuffix(filepath.Base(path), ".go")
		meta, ok := fileMeta[stem]
		if !ok {
			fmt.Fprintf(os.Stderr, "WARN: no metadata for %s, skipping\n", stem)
			continue
		}
		slug, ok := navSlug[stem]
		if !ok {
			slug = stem
		}

		tools, err := extractTools(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			continue
		}

		cat := Category{
			Slug:        slug,
			Title:       meta.title,
			Description: meta.desc,
			Tools:       tools,
		}
		allCategories = append(allCategories, cat)
		totalTools += len(tools)

		outPath := filepath.Join(outDir, slug+".md")
		f, err := os.Create(outPath)
		if err != nil {
			fatal(err)
		}
		if err := mdTemplate.Execute(f, cat); err != nil {
			f.Close()
			fatal(err)
		}
		f.Close()
		fmt.Printf("  ✓ %s → %s (%d tools)\n", filepath.Base(path), outPath, len(tools))
	}

	// Generate index page
	sort.Slice(allCategories, func(i, j int) bool {
		return allCategories[i].Title < allCategories[j].Title
	})
	writeIndex(outDir, allCategories, totalTools)

	fmt.Printf("\n✅ Generated %d category pages with %d total tools\n", len(allCategories), totalTools)
}

func writeIndex(outDir string, cats []Category, total int) {
	f, err := os.Create(filepath.Join(outDir, "index.md"))
	if err != nil {
		fatal(err)
	}
	defer f.Close()

	fmt.Fprintf(f, "# Tool Reference\n\n")
	fmt.Fprintf(f, "The Mythic MCP Server exposes **%d tools** across **%d categories**.\n", total, len(cats))
	fmt.Fprintf(f, "Each tool has a stable name, a description the AI reads to decide when\n")
	fmt.Fprintf(f, "to use it, and a typed parameter schema.\n\n")
	fmt.Fprintf(f, "!!! info \"Auto-generated\"\n")
	fmt.Fprintf(f, "    These pages are generated directly from the Go source code.\n")
	fmt.Fprintf(f, "    They stay in sync with the server automatically on every deploy.\n\n")
	fmt.Fprintf(f, "---\n\n")
	fmt.Fprintf(f, "| Category | Tools | Description |\n")
	fmt.Fprintf(f, "|----------|:-----:|-------------|\n")
	for _, c := range cats {
		fmt.Fprintf(f, "| [%s](%s.md) | %d | %s |\n", c.Title, c.Slug, len(c.Tools), c.Description)
	}
	fmt.Fprintf(f, "\n")
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "FATAL: %v\n", err)
	os.Exit(1)
}
