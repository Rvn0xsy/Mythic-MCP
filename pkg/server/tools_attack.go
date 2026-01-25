package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerAttackTools registers MITRE ATT&CK mapping MCP tools
func (s *Server) registerAttackTools() {
	// mythic_get_attack_techniques - List all MITRE ATT&CK techniques
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_attack_techniques",
		Description: "Get a list of all MITRE ATT&CK techniques available in Mythic",
	}, s.handleGetAttackTechniques)

	// mythic_get_attack_technique_by_id - Get technique by ID
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_attack_technique_by_id",
		Description: "Get details of a specific MITRE ATT&CK technique by internal ID",
	}, s.handleGetAttackTechniqueByID)

	// mythic_get_attack_technique_by_tnum - Get technique by T-number
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_attack_technique_by_tnum",
		Description: "Get details of a specific MITRE ATT&CK technique by T-number (e.g., T1055)",
	}, s.handleGetAttackTechniqueByTNum)

	// mythic_get_attack_by_task - Get techniques used by a task
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_attack_by_task",
		Description: "Get MITRE ATT&CK techniques associated with a specific task",
	}, s.handleGetAttackByTask)

	// mythic_get_attack_by_command - Get techniques for a command
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_attack_by_command",
		Description: "Get MITRE ATT&CK techniques associated with a specific command",
	}, s.handleGetAttackByCommand)

	// mythic_get_attacks_by_operation - Get all techniques used in operation
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_attacks_by_operation",
		Description: "Get all MITRE ATT&CK techniques used in an operation",
	}, s.handleGetAttacksByOperation)
}

// Tool argument types for attack tools

type getAttackTechniquesArgs struct{}

type getAttackTechniqueByIDArgs struct {
	AttackID int `json:"attack_id" jsonschema:"Internal ID of the MITRE ATT&CK technique"`
}

type getAttackTechniqueByTNumArgs struct {
	TNumber string `json:"t_number" jsonschema:"MITRE ATT&CK technique T-number (e.g., T1055)"`
}

type getAttackByTaskArgs struct {
	TaskID int `json:"task_id" jsonschema:"Internal ID of the task"`
}

type getAttackByCommandArgs struct {
	CommandID int `json:"command_id" jsonschema:"Internal ID of the command"`
}

type getAttacksByOperationArgs struct {
	OperationID int `json:"operation_id" jsonschema:"ID of the operation"`
}

// Tool handlers

// handleGetAttackTechniques retrieves all MITRE ATT&CK techniques
func (s *Server) handleGetAttackTechniques(ctx context.Context, req *mcp.CallToolRequest, args getAttackTechniquesArgs) (*mcp.CallToolResult, any, error) {
	techniques, err := s.mythicClient.GetAttackTechniques(ctx)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(techniques, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("All MITRE ATT&CK techniques (%d total):\n\n%s", len(techniques), string(data)),
			},
		},
	}, techniques, nil
}

// handleGetAttackTechniqueByID retrieves a specific technique by ID
func (s *Server) handleGetAttackTechniqueByID(ctx context.Context, req *mcp.CallToolRequest, args getAttackTechniqueByIDArgs) (*mcp.CallToolResult, any, error) {
	technique, err := s.mythicClient.GetAttackTechniqueByID(ctx, args.AttackID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(technique, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("MITRE ATT&CK Technique %s: %s\nTactic: %s\nPlatform: %s\n\n%s",
					technique.TNum, technique.Name, technique.Tactic, technique.OS, string(data)),
			},
		},
	}, technique, nil
}

// handleGetAttackTechniqueByTNum retrieves a technique by T-number
func (s *Server) handleGetAttackTechniqueByTNum(ctx context.Context, req *mcp.CallToolRequest, args getAttackTechniqueByTNumArgs) (*mcp.CallToolResult, any, error) {
	technique, err := s.mythicClient.GetAttackTechniqueByTNum(ctx, args.TNumber)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(technique, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("MITRE ATT&CK Technique %s: %s\nTactic: %s\nPlatform: %s\n\n%s",
					technique.TNum, technique.Name, technique.Tactic, technique.OS, string(data)),
			},
		},
	}, technique, nil
}

// handleGetAttackByTask retrieves techniques associated with a task
func (s *Server) handleGetAttackByTask(ctx context.Context, req *mcp.CallToolRequest, args getAttackByTaskArgs) (*mcp.CallToolResult, any, error) {
	attackTasks, err := s.mythicClient.GetAttackByTask(ctx, args.TaskID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(attackTasks, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("MITRE ATT&CK techniques for task %d (%d total):\n\n%s",
					args.TaskID, len(attackTasks), string(data)),
			},
		},
	}, attackTasks, nil
}

// handleGetAttackByCommand retrieves techniques associated with a command
func (s *Server) handleGetAttackByCommand(ctx context.Context, req *mcp.CallToolRequest, args getAttackByCommandArgs) (*mcp.CallToolResult, any, error) {
	attackCommands, err := s.mythicClient.GetAttackByCommand(ctx, args.CommandID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(attackCommands, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("MITRE ATT&CK techniques for command %d (%d total):\n\n%s",
					args.CommandID, len(attackCommands), string(data)),
			},
		},
	}, attackCommands, nil
}

// handleGetAttacksByOperation retrieves all techniques used in an operation
func (s *Server) handleGetAttacksByOperation(ctx context.Context, req *mcp.CallToolRequest, args getAttacksByOperationArgs) (*mcp.CallToolResult, any, error) {
	techniques, err := s.mythicClient.GetAttacksByOperation(ctx, args.OperationID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(techniques, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	// Build summary by tactic
	tacticMap := make(map[string]int)
	for _, technique := range techniques {
		tacticMap[technique.Tactic]++
	}

	summary := fmt.Sprintf("MITRE ATT&CK techniques used in operation %d (%d total):\n\n", args.OperationID, len(techniques))
	summary += "Coverage by tactic:\n"
	for tactic, count := range tacticMap {
		summary += fmt.Sprintf("  - %s: %d technique(s)\n", tactic, count)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("%s\nFull details:\n\n%s", summary, string(data)),
			},
		},
	}, techniques, nil
}
