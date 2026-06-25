package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"time"
)

type nodeSidecarRequest struct {
	Provider ProviderProfile `json:"provider"`
	Request  GenerateRequest `json:"request"`
	Model    string          `json:"model"`
}

type nodeSidecarResult struct {
	ProviderID string `json:"providerId"`
	Model      string `json:"model"`
	Text       string `json:"text"`
	Error      string `json:"error,omitempty"`
	ErrorClass string `json:"errorClass,omitempty"`
}

func GenerateViaNode(ctx context.Context, provider ProviderProfile, req GenerateRequest, model string) GenerateResult {
	result := GenerateResult{ProviderID: provider.ID, Model: model, BillingMode: string(provider.BillingMode)}
	path := strings.TrimSpace(os.Getenv("HARNEJR_PROVIDER_NODE_PATH"))
	if path == "" {
		result.Error = "node provider runtime requested but HARNEJR_PROVIDER_NODE_PATH is not set"
		result.ErrorClass = "sdk-runtime-missing"
		return result
	}
	payload, err := json.Marshal(nodeSidecarRequest{Provider: provider, Request: req, Model: model})
	if err != nil {
		result.Error = err.Error()
		result.ErrorClass = "payload"
		return result
	}
	timeout := time.Duration(provider.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(callCtx, "node", path, "generate")
	cmd.Stdin = bytes.NewReader(payload)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	start := time.Now()
	err = cmd.Run()
	result.LatencyMs = time.Since(start).Milliseconds()
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		result.Error = message
		result.ErrorClass = "sdk-runtime"
		return result
	}
	var sidecar nodeSidecarResult
	if err := json.Unmarshal(stdout.Bytes(), &sidecar); err != nil {
		result.Error = err.Error()
		result.ErrorClass = "sdk-runtime-parser"
		return result
	}
	if sidecar.Error != "" {
		result.Error = sidecar.Error
		result.ErrorClass = sidecar.ErrorClass
		if result.ErrorClass == "" {
			result.ErrorClass = "sdk-runtime"
		}
		return result
	}
	result.ProviderID = sidecar.ProviderID
	if result.ProviderID == "" {
		result.ProviderID = provider.ID
	}
	result.Model = sidecar.Model
	if result.Model == "" {
		result.Model = model
	}
	result.Text = strings.TrimSpace(sidecar.Text)
	return result
}
