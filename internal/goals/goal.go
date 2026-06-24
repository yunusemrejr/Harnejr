package goals

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type CheckpointStatus string

const (
	CheckpointPending CheckpointStatus = "pending"
	CheckpointDone    CheckpointStatus = "done"
	CheckpointBlocked CheckpointStatus = "blocked"
)

type Checkpoint struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Evidence    string           `json:"evidence"`
	Status      CheckpointStatus `json:"status"`
	CompletedAt string           `json:"completedAt,omitempty"`
	Notes       string           `json:"notes,omitempty"`
}

type State struct {
	Version     int          `json:"version"`
	SessionID   string       `json:"sessionId"`
	Goal        string       `json:"goal"`
	Status      string       `json:"status"`
	StartedAt   string       `json:"startedAt"`
	UpdatedAt   string       `json:"updatedAt"`
	Checkpoints []Checkpoint `json:"checkpoints"`
}

func Start(memoryDir string, sessionID string, goal string) (State, error) {
	goal = strings.TrimSpace(goal)
	if goal == "" {
		return State{}, fmt.Errorf("goal is required")
	}
	now := time.Now().UTC().Format(time.RFC3339)
	state := State{Version: 1, SessionID: strings.TrimSpace(sessionID), Goal: goal, Status: "active", StartedAt: now, UpdatedAt: now, Checkpoints: defaultCheckpoints()}
	return state, save(memoryDir, state)
}

func Load(memoryDir string) (State, error) {
	bytes, err := os.ReadFile(path(memoryDir))
	if os.IsNotExist(err) {
		return State{Version: 1, Status: "none"}, nil
	}
	if err != nil {
		return State{}, err
	}
	var state State
	if err := json.Unmarshal(bytes, &state); err != nil {
		return State{}, err
	}
	return state, nil
}

func Mark(memoryDir string, id string, status CheckpointStatus, notes string) (State, error) {
	state, err := Load(memoryDir)
	if err != nil {
		return State{}, err
	}
	if state.Status == "none" || state.Goal == "" {
		return State{}, fmt.Errorf("no active goal")
	}
	matched := false
	now := time.Now().UTC().Format(time.RFC3339)
	for i := range state.Checkpoints {
		if state.Checkpoints[i].ID == id {
			state.Checkpoints[i].Status = status
			state.Checkpoints[i].Notes = strings.TrimSpace(notes)
			if status == CheckpointDone {
				state.Checkpoints[i].CompletedAt = now
			}
			matched = true
			break
		}
	}
	if !matched {
		return State{}, fmt.Errorf("checkpoint not found: %s", id)
	}
	allDone := true
	for _, checkpoint := range state.Checkpoints {
		if checkpoint.Status != CheckpointDone {
			allDone = false
			break
		}
	}
	if allDone {
		state.Status = "ready_for_completion_review"
	}
	state.UpdatedAt = now
	return state, save(memoryDir, state)
}

func defaultCheckpoints() []Checkpoint {
	return []Checkpoint{
		{ID: "scope", Title: "Understand scope", Evidence: "Workspace, target files, risks, and constraints are identified.", Status: CheckpointPending},
		{ID: "plan", Title: "Plan reversible steps", Evidence: "A small plan exists with rollback-safe changes and verification commands.", Status: CheckpointPending},
		{ID: "implement", Title: "Apply bounded changes", Evidence: "Changes are made through safe file/patch APIs with rollback snapshots.", Status: CheckpointPending},
		{ID: "verify", Title: "Run verification", Evidence: "Relevant tests, builds, linters, or explicit checks were run and recorded.", Status: CheckpointPending},
		{ID: "review", Title: "Independent review", Evidence: "Worker, reviewer, or deterministic completion gate challenged the result.", Status: CheckpointPending},
		{ID: "complete", Title: "Completion decision", Evidence: "Completion is accepted only with evidence, not the main model's claim.", Status: CheckpointPending},
	}
}

func save(memoryDir string, state State) error {
	if err := os.MkdirAll(memoryDir, 0o750); err != nil {
		return err
	}
	bytes, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	bytes = append(bytes, '\n')
	return os.WriteFile(path(memoryDir), bytes, 0o640)
}

func path(memoryDir string) string {
	return filepath.Join(memoryDir, "goal.json")
}
