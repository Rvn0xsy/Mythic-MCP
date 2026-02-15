#!/usr/bin/env python3
"""Enhance handleGetC2Profile to include configuration parameter values."""

with open("pkg/server/tools_c2profiles.go", "r") as f:
    content = f.read()

# 1. Update the tool description for mythic_get_c2_profile
old_desc = '''\t// mythic_get_c2_profile - Get specific C2 profile
\tmcp.AddTool(s.mcpServer, &mcp.Tool{
\t\tName:        "mythic_get_c2_profile",
\t\tDescription: "Get details of a specific C2 profile by ID, including whether it is currently running (listening for callbacks).",
\t}, s.handleGetC2Profile)'''

new_desc = '''\t// mythic_get_c2_profile - Get specific C2 profile with configuration
\tmcp.AddTool(s.mcpServer, &mcp.Tool{
\t\tName:        "mythic_get_c2_profile",
\t\tDescription: "Get details of a specific C2 profile by ID, including its current configuration " +
\t\t\t"parameter values (callback_host, callback_port, etc.) and whether it is currently running.",
\t}, s.handleGetC2Profile)'''

if old_desc not in content:
    print("ERROR: Could not find tool description for mythic_get_c2_profile")
    exit(1)
content = content.replace(old_desc, new_desc)

# 2. Replace the handleGetC2Profile handler to also fetch params
old_handler = '''// handleGetC2Profile retrieves a specific C2 profile by ID
func (s *Server) handleGetC2Profile(ctx context.Context, req *mcp.CallToolRequest, args getC2ProfileArgs) (*mcp.CallToolResult, any, error) {
\tprofile, err := s.mythicClient.GetC2ProfileByID(ctx, args.ProfileID)
\tif err != nil {
\t\treturn nil, nil, translateError(err)
\t}

\tdata, err := json.MarshalIndent(profile, "", "  ")
\tif err != nil {
\t\treturn nil, nil, err
\t}

\tstatus := "Stopped"
\tif profile.Running {
\t\tstatus = "Running"
\t}

\treturn &mcp.CallToolResult{
\t\tContent: []mcp.Content{
\t\t\t&mcp.TextContent{
\t\t\t\tText: fmt.Sprintf("C2 Profile %d (%s): %s\\nType: %s\\n\\n%s",
\t\t\t\t\tprofile.ID, status, profile.Name, profile.Description, string(data)),
\t\t\t},
\t\t},
\t}, profile, nil
}'''

new_handler = '''// handleGetC2Profile retrieves a specific C2 profile by ID, including its
// current configuration parameter values.
func (s *Server) handleGetC2Profile(ctx context.Context, req *mcp.CallToolRequest, args getC2ProfileArgs) (*mcp.CallToolResult, any, error) {
\tprofile, err := s.mythicClient.GetC2ProfileByID(ctx, args.ProfileID)
\tif err != nil {
\t\treturn nil, nil, translateError(err)
\t}

\t// Fetch configuration parameters and populate the profile's Parameters map
\tparams, paramErr := s.mythicClient.GetC2ProfileParameters(ctx, args.ProfileID)
\tif paramErr == nil && len(params) > 0 {
\t\tconfiguration := make(map[string]interface{}, len(params))
\t\tfor _, p := range params {
\t\t\tconfiguration[p.Name] = map[string]interface{}{
\t\t\t\t"value":          p.DefaultValue,
\t\t\t\t"type":           p.ParameterType,
\t\t\t\t"required":       p.Required,
\t\t\t\t"description":    p.Description,
\t\t\t\t"verifier_regex": p.VerifierRegex,
\t\t\t}
\t\t\t// Include choices for selection types
\t\t\tif p.Choices != "" && p.Choices != "[]" {
\t\t\t\tconfiguration[p.Name].(map[string]interface{})["choices"] = p.Choices
\t\t\t}
\t\t}
\t\tprofile.Parameters = configuration
\t}

\tdata, err := json.MarshalIndent(profile, "", "  ")
\tif err != nil {
\t\treturn nil, nil, err
\t}

\tstatus := "Stopped"
\tif profile.Running {
\t\tstatus = "Running"
\t}

\treturn &mcp.CallToolResult{
\t\tContent: []mcp.Content{
\t\t\t&mcp.TextContent{
\t\t\t\tText: fmt.Sprintf("C2 Profile %d (%s): %s\\nType: %s\\n\\n%s",
\t\t\t\t\tprofile.ID, status, profile.Name, profile.Description, string(data)),
\t\t\t},
\t\t},
\t}, profile, nil
}'''

if old_handler not in content:
    print("ERROR: Could not find handleGetC2Profile handler")
    exit(1)
content = content.replace(old_handler, new_handler)

with open("pkg/server/tools_c2profiles.go", "w") as f:
    f.write(content)

print("OK: handleGetC2Profile now includes configuration parameters")
