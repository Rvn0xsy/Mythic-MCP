package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/nbaertsch/mythic-sdk-go/pkg/mythic"
	"github.com/nbaertsch/mythic-sdk-go/pkg/mythic/types"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()

	mythicURL := envOrDefault("MYTHIC_URL", "https://127.0.0.1:7443")
	mythicUsername := envOrDefault("MYTHIC_USERNAME", "mythic_admin")
	mythicPassword := os.Getenv("MYTHIC_PASSWORD")
	if mythicPassword == "" {
		fatalf("MYTHIC_PASSWORD is required")
	}

	client, err := mythic.NewClient(&mythic.Config{
		ServerURL:     mythicURL,
		Username:      mythicUsername,
		Password:      mythicPassword,
		SSL:           true,
		SkipTLSVerify: true,
		Timeout:       30 * time.Second,
	})
	if err != nil {
		fatalf("failed to create mythic client: %v", err)
	}
	if err := client.Login(ctx); err != nil {
		fatalf("failed to login to mythic: %v", err)
	}

	if err := ensureCurrentOperation(ctx, client); err != nil {
		fatalf("failed to ensure current operation: %v", err)
	}

	payloadTypeName, selectedOS, err := choosePayloadTypeAndOS(ctx, client)
	if err != nil {
		fatalf("failed to choose payload type: %v", err)
	}

	payloadType, err := payloadTypeByName(ctx, client, payloadTypeName)
	if err != nil {
		fatalf("failed to lookup payload type %q: %v", payloadTypeName, err)
	}

	payloadUUID, payloadPath, err := buildAndDownloadPayload(ctx, client, payloadTypeName, selectedOS)
	if err != nil {
		fatalf("failed to build/download payload: %v", err)
	}

	baselineMax, err := maxCallbackDisplayID(ctx, client)
	if err != nil {
		fatalf("failed to query baseline callbacks: %v", err)
	}

	cb1, err := startAndWaitForNewCallback(ctx, client, payloadPath, baselineMax)
	if err != nil {
		fatalf("failed to start first callback: %v", err)
	}
	cb2, err := startAndWaitForNewCallback(ctx, client, payloadPath, cb1.DisplayID)
	if err != nil {
		fatalf("failed to start second callback: %v", err)
	}

	seededScreenshot := false
	seededKeylog := false

	commands, err := client.GetCommandsByPayloadType(ctx, payloadType.ID)
	if err != nil {
		fatalf("failed to query commands for payload type %s: %v", payloadTypeName, err)
	}

	screenshotCmd := firstCommand(commands, []string{"screencapture", "screenshot"})
	keylogCmd := firstCommand(commands, []string{"keylog", "keylogger", "keylogger_start"})

	// Try to generate artifacts on cb1.
	if screenshotCmd != "" {
		if err := issueAndWaitTask(ctx, client, cb1.DisplayID, screenshotCmd, ""); err != nil {
			fmt.Fprintf(os.Stderr, "WARN: screenshot task failed: %v\n", err)
		} else {
			seededScreenshot = waitForScreenshot(ctx, client, cb1.DisplayID, 60*time.Second)
		}
	} else {
		fmt.Fprintf(os.Stderr, "WARN: no screenshot command found for payload type %s\n", payloadTypeName)
	}

	if keylogCmd != "" {
		if err := issueAndWaitTask(ctx, client, cb1.DisplayID, keylogCmd, ""); err != nil {
			fmt.Fprintf(os.Stderr, "WARN: keylog task failed: %v\n", err)
		} else {
			seededKeylog = waitForKeylog(ctx, client, cb1.DisplayID, 60*time.Second)
		}
	} else {
		fmt.Fprintf(os.Stderr, "WARN: no keylog command found for payload type %s\n", payloadTypeName)
	}

	writeOutput("payload_type", payloadTypeName)
	writeOutput("payload_uuid", payloadUUID)
	writeOutput("callback_display_id_1", strconv.Itoa(cb1.DisplayID))
	writeOutput("callback_display_id_2", strconv.Itoa(cb2.DisplayID))
	writeOutput("seeded_screenshot", boolString(seededScreenshot))
	writeOutput("seeded_keylog", boolString(seededKeylog))

	fmt.Printf("seeded payload_type=%s payload_uuid=%s callbacks=[%d,%d] screenshot=%v keylog=%v\n", payloadTypeName, payloadUUID, cb1.DisplayID, cb2.DisplayID, seededScreenshot, seededKeylog)
}

func envOrDefault(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	os.Exit(1)
}

func boolString(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func writeOutput(key, value string) {
	path := os.Getenv("GITHUB_OUTPUT")
	if path == "" {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARN: failed to open GITHUB_OUTPUT: %v\n", err)
		return
	}
	defer f.Close()
	_, _ = fmt.Fprintf(f, "%s=%s\n", key, value)
}

func ensureCurrentOperation(ctx context.Context, client *mythic.Client) error {
	me, err := client.GetMe(ctx)
	if err != nil {
		return err
	}
	if me.CurrentOperation != nil {
		return nil
	}
	ops, err := client.GetOperations(ctx)
	if err != nil {
		return err
	}
	if len(ops) == 0 {
		return errors.New("no operations found")
	}
	return client.UpdateCurrentOperationForUser(ctx, ops[0].ID)
}

func choosePayloadTypeAndOS(ctx context.Context, client *mythic.Client) (payloadTypeName string, selectedOS string, err error) {
	payloadTypes, err := client.GetPayloadTypes(ctx)
	if err != nil {
		return "", "", err
	}
	available := map[string]bool{}
	for _, pt := range payloadTypes {
		available[strings.ToLower(pt.Name)] = true
	}

	// Prefer Apollo because it can plausibly run on Linux runners.
	if available["apollo"] {
		return "apollo", normalizeOS(runtime.GOOS), nil
	}
	// Poseidon is macOS-focused; only pick it if we are on darwin.
	if available["poseidon"] && runtime.GOOS == "darwin" {
		return "poseidon", "macOS", nil
	}
	return "", "", fmt.Errorf("no suitable payload type found (need apollo, or poseidon on darwin)")
}

func normalizeOS(goos string) string {
	switch goos {
	case "linux":
		return "Linux"
	case "darwin":
		return "macOS"
	case "windows":
		return "Windows"
	default:
		return "Linux"
	}
}

func payloadTypeByName(ctx context.Context, client *mythic.Client, name string) (*types.PayloadType, error) {
	pts, err := client.GetPayloadTypes(ctx)
	if err != nil {
		return nil, err
	}
	for _, pt := range pts {
		if strings.EqualFold(pt.Name, name) {
			return pt, nil
		}
	}
	return nil, fmt.Errorf("payload type %q not found", name)
}

func buildAndDownloadPayload(ctx context.Context, client *mythic.Client, payloadTypeName string, selectedOS string) (uuid string, path string, err error) {
	// Determine build parameters; populate required params with defaults when possible.
	payloadType, err := payloadTypeByName(ctx, client, payloadTypeName)
	if err != nil {
		return "", "", err
	}
	params, err := client.GetBuildParametersByPayloadType(ctx, payloadType.ID)
	if err != nil {
		return "", "", err
	}
	buildParams := map[string]interface{}{}
	for _, p := range params {
		if !p.Required {
			continue
		}
		if p.DefaultValue != "" {
			buildParams[p.Name] = p.DefaultValue
			continue
		}
		// Best-effort defaults for common required fields.
		nameLower := strings.ToLower(p.Name)
		switch {
		case strings.Contains(nameLower, "arch"):
			buildParams[p.Name] = "x64"
		case strings.Contains(nameLower, "debug"):
			buildParams[p.Name] = false
		default:
			// Leave unset; Mythic may still accept empty.
		}
	}

	req := &types.CreatePayloadRequest{
		PayloadType: payloadTypeName,
		SelectedOS:  selectedOS,
		Filename:    fmt.Sprintf("e2e-seed-%s", payloadTypeName),
		Description: "E2E seed payload",
		C2Profiles: []types.C2ProfileConfig{
			{
				Name: "http",
				Parameters: map[string]interface{}{
					"callback_host": "https://127.0.0.1:7443",
				},
			},
		},
		BuildParameters: buildParams,
	}

	payload, err := client.CreatePayload(ctx, req)
	if err != nil {
		return "", "", err
	}

	if err := client.WaitForPayloadComplete(ctx, payload.UUID, 420); err != nil {
		return payload.UUID, "", err
	}

	data, err := client.DownloadPayload(ctx, payload.UUID)
	if err != nil {
		return payload.UUID, "", err
	}

	outPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%s", payloadTypeName, payload.UUID))
	if err := os.WriteFile(outPath, data, 0o700); err != nil {
		return payload.UUID, "", err
	}

	return payload.UUID, outPath, nil
}

func maxCallbackDisplayID(ctx context.Context, client *mythic.Client) (int, error) {
	cbs, err := client.GetAllCallbacks(ctx)
	if err != nil {
		return 0, err
	}
	max := 0
	for _, cb := range cbs {
		if cb.DisplayID > max {
			max = cb.DisplayID
		}
	}
	return max, nil
}

func startAndWaitForNewCallback(ctx context.Context, client *mythic.Client, payloadPath string, baselineMax int) (*types.Callback, error) {
	pid, err := startDetached(payloadPath)
	if err != nil {
		return nil, err
	}
	fmt.Printf("started payload pid=%d baseline_max=%d\n", pid, baselineMax)

	deadline := time.Now().Add(180 * time.Second)
	for time.Now().Before(deadline) {
		cbs, err := client.GetAllActiveCallbacks(ctx)
		if err != nil {
			return nil, err
		}
		for _, cb := range cbs {
			if cb.DisplayID > baselineMax {
				fmt.Printf("new callback: display_id=%d user=%s host=%s\n", cb.DisplayID, cb.User, cb.Host)
				return cb, nil
			}
		}
		time.Sleep(3 * time.Second)
	}
	return nil, fmt.Errorf("timeout waiting for callback > %d", baselineMax)
}

func startDetached(payloadPath string) (int, error) {
	logPath := filepath.Join(os.TempDir(), fmt.Sprintf("e2e-seed-%d.log", time.Now().UnixNano()))
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return 0, err
	}

	cmd := exec.Command(payloadPath)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return 0, err
	}
	// Best-effort close our copy; child keeps fd.
	_ = logFile.Close()
	return cmd.Process.Pid, nil
}

func firstCommand(cmds []*types.Command, preferred []string) string {
	set := map[string]bool{}
	for _, c := range cmds {
		set[strings.ToLower(c.Cmd)] = true
	}
	for _, p := range preferred {
		if set[strings.ToLower(p)] {
			return p
		}
	}
	return ""
}

func issueAndWaitTask(ctx context.Context, client *mythic.Client, callbackDisplayID int, command string, params string) error {
	req := &mythic.TaskRequest{
		CallbackID:        &callbackDisplayID,
		Command:           command,
		Params:            params,
		IsInteractiveTask: false,
	}
	task, err := client.IssueTask(ctx, req)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("issue task returned nil")
	}
	return client.WaitForTaskComplete(ctx, task.DisplayID, 120)
}

func waitForScreenshot(ctx context.Context, client *mythic.Client, callbackDisplayID int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		screens, err := client.GetScreenshots(ctx, callbackDisplayID, 10)
		if err == nil && len(screens) > 0 {
			return true
		}
		time.Sleep(3 * time.Second)
	}
	return false
}

func waitForKeylog(ctx context.Context, client *mythic.Client, callbackDisplayID int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		kls, err := client.GetKeylogsByCallback(ctx, callbackDisplayID)
		if err == nil && len(kls) > 0 {
			return true
		}
		time.Sleep(3 * time.Second)
	}
	return false
}
