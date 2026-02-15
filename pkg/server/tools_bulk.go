package server

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nbaertsch/mythic-sdk-go/pkg/mythic"
)

const maxBulkCallbacks = 50

// registerBulkTools registers bulk operation MCP tools
func (s *Server) registerBulkTools() {
	// mythic_issue_task_bulk - Issue the same command to multiple callbacks
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name: "mythic_issue_task_bulk",
		Description: "Issue the same command to multiple callbacks in parallel. Useful for " +
			"running enumeration commands (ps, ifconfig, ls) across many agents simultaneously. " +
			"Returns results for each callback including successes and failures. " +
			"Maximum " + fmt.Sprintf("%d", maxBulkCallbacks) + " callbacks per call.",
	}, s.handleIssueTaskBulk)

	// mythic_get_tasks_batch - Retrieve multiple tasks by display IDs
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name: "mythic_get_tasks_batch",
		Description: "Retrieve multiple tasks by their display IDs in a single call. " +
			"Efficient alternative to calling mythic_get_task repeatedly. " +
			"Tasks that don't exist are silently omitted from the results.",
	}, s.handleGetTasksBatch)

	// mythic_search_tasks - Search/filter tasks across callbacks
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name: "mythic_search_tasks",
		Description: "Search for tasks across all callbacks with flexible filters. " +
			"Filter by command name, status, callback, or any combination. " +
			"Useful for reviewing operations, finding specific command executions, " +
			"or checking completion status of multiple tasks.",
	}, s.handleSearchTasks)
}

// Argument types for bulk tools

type issueTaskBulkArgs struct {
	CallbackIDs []int  `json:"callback_ids" jsonschema:"List of callback display_ids to task. Get these from mythic_get_callbacks or mythic_get_active_callbacks. Maximum 50."`
	Command     string `json:"command" jsonschema:"Command name to execute on all callbacks (e.g. ps, ifconfig, ls)."`
	Params      string `json:"params,omitempty" jsonschema:"Command parameters (same for all callbacks). See mythic_issue_task for format details."`
	PayloadType string `json:"payload_type,omitempty" jsonschema:"Optional: payload type for cross-type tasking."`
}

type getTasksBatchArgs struct {
	TaskIDs []int `json:"task_ids" jsonschema:"List of task display_ids to retrieve."`
}

type searchTasksArgs struct {
	Command    string `json:"command,omitempty" jsonschema:"Filter by command name (exact match, e.g. 'shell', 'ps')."`
	Status     string `json:"status,omitempty" jsonschema:"Filter by task status: submitted, processing, processed, completed, error."`
	CallbackID int    `json:"callback_id,omitempty" jsonschema:"Filter by callback display_id. If omitted, searches across all callbacks."`
	Limit      int    `json:"limit,omitempty" jsonschema:"Maximum number of tasks to return (default 50, max 200)."`
}

// Bulk task result for a single callback
type bulkTaskResult struct {
	CallbackID int          `json:"callback_id"`
	Task       *mythic.Task `json:"task,omitempty"`
	Error      string       `json:"error,omitempty"`
	Success    bool         `json:"success"`
}

// handleIssueTaskBulk issues the same command to multiple callbacks concurrently
func (s *Server) handleIssueTaskBulk(ctx context.Context, req *mcp.CallToolRequest, args issueTaskBulkArgs) (*mcp.CallToolResult, any, error) {
	if len(args.CallbackIDs) == 0 {
		return nil, nil, fmt.Errorf("at least one callback_id is required")
	}
	if len(args.CallbackIDs) > maxBulkCallbacks {
		return nil, nil, fmt.Errorf("too many callbacks: %d exceeds maximum of %d", len(args.CallbackIDs), maxBulkCallbacks)
	}

	// Issue tasks concurrently
	results := make([]bulkTaskResult, len(args.CallbackIDs))
	var wg sync.WaitGroup

	for i, cbID := range args.CallbackIDs {
		wg.Add(1)
		go func(idx int, callbackID int) {
			defer wg.Done()

			// Smart parameter resolution per callback
			var payloadTypeForResolve *string
			if args.PayloadType != "" {
				payloadTypeForResolve = &args.PayloadType
			}
			params := s.resolveTaskParams(ctx, callbackID, args.Command, args.Params, payloadTypeForResolve)

			issueReq := &mythic.TaskRequest{
				CallbackID: &callbackID,
				Command:    args.Command,
				Params:     params,
			}
			if args.PayloadType != "" {
				pt := args.PayloadType
				issueReq.PayloadType = &pt
			}

			task, err := s.mythicClient.IssueTask(ctx, issueReq)
			if err != nil {
				results[idx] = bulkTaskResult{
					CallbackID: callbackID,
					Error:      translateError(err).Error(),
					Success:    false,
				}
				return
			}

			results[idx] = bulkTaskResult{
				CallbackID: callbackID,
				Task:       task,
				Success:    true,
			}
		}(i, cbID)
	}

	wg.Wait()

	// Summarize results
	var succeeded, failed int
	for _, r := range results {
		if r.Success {
			succeeded++
		} else {
			failed++
		}
	}

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	summary := fmt.Sprintf("Bulk task '%s' issued to %d callbacks: %d succeeded, %d failed\n\n%s",
		args.Command, len(args.CallbackIDs), succeeded, failed, string(data))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
		},
	}, results, nil
}

// handleGetTasksBatch retrieves multiple tasks by display IDs in a single query
func (s *Server) handleGetTasksBatch(ctx context.Context, req *mcp.CallToolRequest, args getTasksBatchArgs) (*mcp.CallToolResult, any, error) {
	if len(args.TaskIDs) == 0 {
		return nil, nil, fmt.Errorf("at least one task_id is required")
	}

	tasks, err := s.mythicClient.GetTasksByDisplayIDs(ctx, args.TaskIDs)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	// Report which IDs were not found
	foundIDs := make(map[int]bool)
	for _, t := range tasks {
		foundIDs[t.DisplayID] = true
	}
	var missing []int
	for _, id := range args.TaskIDs {
		if !foundIDs[id] {
			missing = append(missing, id)
		}
	}

	summary := fmt.Sprintf("Retrieved %d of %d requested tasks", len(tasks), len(args.TaskIDs))
	if len(missing) > 0 {
		summary += fmt.Sprintf(" (not found: %v)", missing)
	}
	summary += fmt.Sprintf("\n\n%s", string(data))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
		},
	}, wrapList(tasks), nil
}

// handleSearchTasks searches for tasks with flexible filters
func (s *Server) handleSearchTasks(ctx context.Context, req *mcp.CallToolRequest, args searchTasksArgs) (*mcp.CallToolResult, any, error) {
	if args.Command == "" && args.Status == "" && args.CallbackID == 0 {
		return nil, nil, fmt.Errorf("at least one filter is required: command, status, or callback_id")
	}

	limit := args.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	// Use GraphQL directly for flexible filtering
	tasks, err := s.searchTasksGraphQL(ctx, args.Command, args.Status, args.CallbackID, limit)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	filterDesc := "Filters:"
	if args.Command != "" {
		filterDesc += fmt.Sprintf(" command=%s", args.Command)
	}
	if args.Status != "" {
		filterDesc += fmt.Sprintf(" status=%s", args.Status)
	}
	if args.CallbackID > 0 {
		filterDesc += fmt.Sprintf(" callback=%d", args.CallbackID)
	}

	summary := fmt.Sprintf("Found %d tasks matching filters\n%s\n\n%s", len(tasks), filterDesc, string(data))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
		},
	}, wrapList(tasks), nil
}

// searchTasksGraphQL builds a dynamic GraphQL query based on the provided filters.
// This is implemented at the MCP layer rather than the SDK because the hasura-style
// dynamic where clause composition doesn't map cleanly to the SDK's typed approach.
func (s *Server) searchTasksGraphQL(ctx context.Context, command, status string, callbackDisplayID, limit int) ([]map[string]interface{}, error) {
	// Build the where clause dynamically
	where := make(map[string]interface{})

	if command != "" {
		where["command_name"] = map[string]interface{}{"_eq": command}
	}
	if status != "" {
		where["status"] = map[string]interface{}{"_eq": status}
	}
	if callbackDisplayID > 0 {
		// Need to resolve display_id to internal callback_id
		callback, err := s.mythicClient.GetCallbackByID(ctx, callbackDisplayID)
		if err != nil {
			return nil, err
		}
		where["callback_id"] = map[string]interface{}{"_eq": callback.ID}
	}

	// Use the SDK's raw query method
	query := `query SearchTasks($where: task_bool_exp!, $limit: Int!) {
		task(where: $where, order_by: {id: desc}, limit: $limit) {
			id
			display_id
			command_name
			display_params
			original_params
			status
			completed
			comment
			timestamp
			callback_id
			operator_id
			response_count
			callback {
				display_id
			}
		}
	}`

	variables := map[string]interface{}{
		"where": where,
		"limit": limit,
	}

	result, err := s.mythicClient.ExecuteRawGraphQL(ctx, query, variables)
	if err != nil {
		return nil, err
	}

	// Extract the task array from the result
	taskData, ok := result["task"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format: missing task array")
	}

	tasks := make([]map[string]interface{}, 0, len(taskData))
	for _, item := range taskData {
		task, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		// Enrich with callback display_id for convenience
		if cb, ok := task["callback"].(map[string]interface{}); ok {
			task["callback_display_id"] = cb["display_id"]
		}
		delete(task, "callback") // Remove nested object
		tasks = append(tasks, task)
	}

	return tasks, nil
}
