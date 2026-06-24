package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Mode string

const (
	ModeIdle Mode = "idle"
	ModeGoal Mode = "goal"
	ModeLoop Mode = "loop"
)

type ControlState struct {
	SessionID string    `json:"sessionId"`
	Mode      Mode      `json:"mode"`
	Topic     string    `json:"topic,omitempty"`
	Goal      string    `json:"goal,omitempty"`
	LoopTask  string    `json:"loopTask,omitempty"`
	LoopLimit int       `json:"loopLimit,omitempty"`
	Yolo      bool      `json:"yolo"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ControlPatch struct {
	SessionID string `json:"sessionId"`
	Mode      Mode   `json:"mode"`
	Topic     string `json:"topic"`
	Goal      string `json:"goal"`
	LoopTask  string `json:"loopTask"`
	LoopLimit int    `json:"loopLimit"`
	Yolo      *bool  `json:"yolo"`
	ClearGoal bool   `json:"clearGoal"`
	ClearTopic bool   `json:"clearTopic"`
}

func LoadControl(memoryDir string) (ControlState, error) {
	path := controlPath(memoryDir)
	bytes, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return ControlState{Mode: ModeIdle}, nil
	}
	if err != nil {
		return ControlState{}, err
	}
	var state ControlState
	if err := json.Unmarshal(bytes, &state); err != nil {
		return ControlState{}, err
	}
	if state.Mode == "" {
		state.Mode = ModeIdle
	}
	return state, nil
}

func ApplyControl(memoryDir string, patch ControlPatch, now time.Time) (ControlState, error) {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	state, err := LoadControl(memoryDir)
	if err != nil {
		return ControlState{}, err
	}
	if patch.SessionID != "" {
		state.SessionID = strings.TrimSpace(patch.SessionID)
	}
	if patch.ClearTopic {
		state.Topic = ""
	} else if strings.TrimSpace(patch.Topic) != "" {
		state.Topic = strings.TrimSpace(patch.Topic)
	}
	if patch.ClearGoal {
		state.Goal = ""
		if state.Mode == ModeGoal {
			state.Mode = ModeIdle
		}
	} else if strings.TrimSpace(patch.Goal) != "" {
		if state.Mode == ModeLoop {
			return ControlState{}, fmt.Errorf("goal mode cannot start while loop mode is active")
		}
		state.Goal = strings.TrimSpace(patch.Goal)
		state.Mode = ModeGoal
	}
	if strings.TrimSpace(patch.LoopTask) != "" || patch.Mode == ModeLoop {
		if state.Mode == ModeGoal && state.Goal != "" {
			return ControlState{}, fmt.Errorf("loop mode cannot start while a goal is active")
		}
		state.Mode = ModeLoop
		state.LoopTask = strings.TrimSpace(patch.LoopTask)
		state.LoopLimit = patch.LoopLimit
		if state.LoopLimit < 1 {
			state.LoopLimit = 1
		}
	}
	if patch.Mode == ModeIdle {
		state.Mode = ModeIdle
		state.LoopTask = ""
		state.LoopLimit = 0
	}
	if patch.Yolo != nil {
		state.Yolo = *patch.Yolo
	}
	state.UpdatedAt = now.UTC()
	if err := os.MkdirAll(memoryDir, 0o750); err != nil {
		return ControlState{}, err
	}
	bytes, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return ControlState{}, err
	}
	if err := os.WriteFile(controlPath(memoryDir), append(bytes, '\n'), 0o640); err != nil {
		return ControlState{}, err
	}
	return state, nil
}

func controlPath(memoryDir string) string {
	return filepath.Join(memoryDir, "control.json")
}
