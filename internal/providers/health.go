package providers

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const healthVersion = 1

var healthMu sync.Mutex

type ProviderHealthState string

const (
	HealthUnknown ProviderHealthState = "unknown"
	HealthHealthy ProviderHealthState = "healthy"
	HealthDegraded ProviderHealthState = "degraded"
	HealthCooling ProviderHealthState = "cooling_down"
	HealthDown ProviderHealthState = "down"
)

type ProviderHealthEntry struct {
	ProviderID    string              `json:"providerId"`
	State         ProviderHealthState `json:"state"`
	Failures      int                 `json:"failures"`
	Successes     int                 `json:"successes"`
	LastError     string              `json:"lastError,omitempty"`
	LastErrorClass string             `json:"lastErrorClass,omitempty"`
	LastStatusCode int                `json:"lastStatusCode,omitempty"`
	LastLatencyMs int64               `json:"lastLatencyMs,omitempty"`
	CooldownUntil string              `json:"cooldownUntil,omitempty"`
	UpdatedAt     string             `json:"updatedAt,omitempty"`
}

type HealthLedger struct {
	Version   int                            `json:"version"`
	UpdatedAt string                         `json:"updatedAt,omitempty"`
	Providers map[string]ProviderHealthEntry `json:"providers"`
}

func NewHealthLedger() HealthLedger {
	return HealthLedger{Version: healthVersion, Providers: map[string]ProviderHealthEntry{}}
}

func LoadHealthLedger(path string) (HealthLedger, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewHealthLedger(), nil
		}
		return NewHealthLedger(), err
	}
	var ledger HealthLedger
	if err := json.Unmarshal(data, &ledger); err != nil {
		return NewHealthLedger(), err
	}
	if ledger.Providers == nil {
		ledger.Providers = map[string]ProviderHealthEntry{}
	}
	if ledger.Version == 0 {
		ledger.Version = healthVersion
	}
	return ledger, nil
}

func SaveHealthLedger(path string, ledger HealthLedger) error {
	if ledger.Providers == nil {
		ledger.Providers = map[string]ProviderHealthEntry{}
	}
	ledger.Version = healthVersion
	ledger.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(ledger, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

func ResetHealth(path string, providerID string) (HealthLedger, error) {
	healthMu.Lock()
	defer healthMu.Unlock()
	ledger, err := LoadHealthLedger(path)
	if err != nil {
		return ledger, err
	}
	providerID = strings.TrimSpace(providerID)
	if providerID == "" {
		ledger.Providers = map[string]ProviderHealthEntry{}
	} else {
		delete(ledger.Providers, providerID)
	}
	return ledger, SaveHealthLedger(path, ledger)
}

func (ledger HealthLedger) Entry(providerID string) ProviderHealthEntry {
	if entry, ok := ledger.Providers[providerID]; ok {
		return entry
	}
	return ProviderHealthEntry{ProviderID: providerID, State: HealthUnknown}
}

func (ledger HealthLedger) CanUse(providerID string, now time.Time) (bool, string) {
	entry := ledger.Entry(providerID)
	switch entry.State {
	case HealthDown:
		return false, "provider marked down: " + entry.LastErrorClass
	case HealthCooling:
		if until, ok := parseHealthTime(entry.CooldownUntil); ok && now.Before(until) {
			return false, "provider cooling down until " + entry.CooldownUntil
		}
	}
	return true, ""
}

func (ledger *HealthLedger) RecordResult(result GenerateResult, now time.Time) {
	if strings.TrimSpace(result.ProviderID) == "" {
		return
	}
	entry := ledger.Entry(result.ProviderID)
	entry.ProviderID = result.ProviderID
	entry.UpdatedAt = now.UTC().Format(time.RFC3339)
	entry.LastStatusCode = result.StatusCode
	entry.LastLatencyMs = result.LatencyMs
	if result.Error == "" && strings.TrimSpace(result.Text) != "" {
		entry.State = HealthHealthy
		entry.Failures = 0
		entry.Successes++
		entry.LastError = ""
		entry.LastErrorClass = ""
		entry.CooldownUntil = ""
		ledger.Providers[result.ProviderID] = entry
		return
	}
	entry.Failures++
	entry.LastError = strings.TrimSpace(result.Error)
	entry.LastErrorClass = strings.TrimSpace(result.ErrorClass)
	entry.State = classifyHealthState(entry.LastErrorClass, entry.Failures)
	entry.CooldownUntil = cooldownUntil(entry.LastErrorClass, entry.Failures, now)
	ledger.Providers[result.ProviderID] = entry
}

func GenerateWithHealth(ctx context.Context, registry Registry, healthPath string, preferred []string, req GenerateRequest) GenerateResult {
	candidates := generationCandidates(registry, preferred)
	if len(candidates) == 0 {
		return GenerateResult{Error: "no provider candidate available", ErrorClass: "no-provider"}
	}
	now := time.Now().UTC()
	ledger := NewHealthLedger()
	if strings.TrimSpace(healthPath) != "" {
		healthMu.Lock()
		loaded, err := LoadHealthLedger(healthPath)
		if err == nil {
			ledger = loaded
		}
		healthMu.Unlock()
	}
	usable := make([]ProviderProfile, 0, len(candidates))
	var skipped []string
	for _, provider := range candidates {
		if ok, reason := ledger.CanUse(provider.ID, now); ok {
			usable = append(usable, provider)
		} else {
			skipped = append(skipped, provider.ID+": "+reason)
		}
	}
	if len(usable) == 0 {
		return GenerateResult{Error: "no provider candidate available after health filtering", ErrorClass: "provider-health", HealthSkipped: skipped}
	}
	baseBilling := usable[0].BillingMode
	var tried []string
	var last GenerateResult
	for index, provider := range usable {
		if index > 0 && !req.AllowBillingChange && provider.BillingMode != baseBilling {
			last = GenerateResult{ProviderID: provider.ID, Model: provider.DefaultModel, BillingMode: string(provider.BillingMode), Tried: tried, HealthSkipped: skipped, Error: "fallback would change billing mode from " + string(baseBilling) + " to " + string(provider.BillingMode), ErrorClass: "billing-change", BillingChangeRequired: true}
			continue
		}
		tried = append(tried, provider.ID)
		result := Generate(ctx, provider, req)
		result.Tried = append([]string{}, tried...)
		result.HealthSkipped = append([]string{}, skipped...)
		if strings.TrimSpace(healthPath) != "" {
			healthMu.Lock()
			ledger, err := LoadHealthLedger(healthPath)
			if err == nil {
				ledger.RecordResult(result, time.Now().UTC())
				_ = SaveHealthLedger(healthPath, ledger)
			}
			healthMu.Unlock()
		}
		if result.Error == "" && strings.TrimSpace(result.Text) != "" {
			return result
		}
		last = result
		if !retryableGenerationError(result.ErrorClass) {
			continue
		}
	}
	if last.Error != "" {
		last.Tried = tried
		last.HealthSkipped = skipped
		return last
	}
	return GenerateResult{Tried: tried, HealthSkipped: skipped, Error: "no provider candidate succeeded", ErrorClass: "unknown"}
}

func classifyHealthState(errorClass string, failures int) ProviderHealthState {
	switch errorClass {
	case "auth", "wrong-endpoint-or-model", "unsupported-field", "unsupported-protocol":
		return HealthDown
	case "rate-limit-or-quota":
		return HealthCooling
	case "network", "timeout", "server-retryable", "stream", "sdk-runtime":
		if failures >= 2 {
			return HealthCooling
		}
		return HealthDegraded
	case "parser", "payload", "context-too-large":
		return HealthDegraded
	default:
		if failures >= 3 {
			return HealthCooling
		}
		return HealthDegraded
	}
}

func cooldownUntil(errorClass string, failures int, now time.Time) string {
	switch errorClass {
	case "rate-limit-or-quota":
		return now.Add(15 * time.Minute).UTC().Format(time.RFC3339)
	case "network", "timeout", "server-retryable", "stream", "sdk-runtime":
		if failures >= 2 {
			return now.Add(2 * time.Minute).UTC().Format(time.RFC3339)
		}
	}
	return ""
}

func parseHealthTime(value string) (time.Time, bool) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339, value)
	return parsed, err == nil
}
