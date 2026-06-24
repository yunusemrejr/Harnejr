package goals

import "testing"

func TestGoalLifecycle(t *testing.T) {
	dir := t.TempDir()
	state, err := Start(dir, "s1", "finish task")
	if err != nil {
		t.Fatal(err)
	}
	if state.Status != "active" || len(state.Checkpoints) == 0 {
		t.Fatalf("unexpected state: %#v", state)
	}
	updated, err := Mark(dir, "scope", CheckpointDone, "ok")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Checkpoints[0].Status != CheckpointDone {
		t.Fatalf("checkpoint not marked: %#v", updated.Checkpoints[0])
	}
}
