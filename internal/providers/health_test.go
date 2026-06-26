package providers

import (
	"path/filepath"
	"testing"
	"time"
)

func TestHealthLedgerRecordsCooldownAndRecovery(t *testing.T) {
	now := time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC)
	ledger := NewHealthLedger()
	ledger.RecordResult(GenerateResult{ProviderID: "stepfun-step-plan", Error: "quota exceeded", ErrorClass: "rate-limit-or-quota", StatusCode: 429}, now)

	entry := ledger.Entry("stepfun-step-plan")
	if entry.State != HealthCooling {
		t.Fatalf("expected cooling_down, got %s", entry.State)
	}
	if ok, _ := ledger.CanUse("stepfun-step-plan", now.Add(time.Minute)); ok {
		t.Fatalf("provider should be filtered during cooldown")
	}
	if ok, reason := ledger.CanUse("stepfun-step-plan", now.Add(16*time.Minute)); !ok {
		t.Fatalf("provider should recover after cooldown, reason: %s", reason)
	}

	ledger.RecordResult(GenerateResult{ProviderID: "stepfun-step-plan", Text: "valid verifier output", LatencyMs: 12}, now.Add(17*time.Minute))
	entry = ledger.Entry("stepfun-step-plan")
	if entry.State != HealthHealthy || entry.Failures != 0 || entry.CooldownUntil != "" {
		t.Fatalf("success should reset failures and cooldown: %+v", entry)
	}
}

func TestHealthLedgerPersistsAndResets(t *testing.T) {
	path := filepath.Join(t.TempDir(), "provider-health.json")
	ledger := NewHealthLedger()
	ledger.RecordResult(GenerateResult{ProviderID: "nvidia-build-nim", Error: "401", ErrorClass: "auth", StatusCode: 401}, time.Now().UTC())
	if err := SaveHealthLedger(path, ledger); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadHealthLedger(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Entry("nvidia-build-nim").State != HealthDown {
		t.Fatalf("expected persisted down state, got %+v", loaded.Entry("nvidia-build-nim"))
	}
	reset, err := ResetHealth(path, "nvidia-build-nim")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := reset.Providers["nvidia-build-nim"]; ok {
		t.Fatalf("provider health entry should have been removed")
	}
}
