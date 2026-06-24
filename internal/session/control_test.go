package session

import "testing"

func TestApplyControlPreventsGoalLoopConflict(t *testing.T) {
	dir := t.TempDir()
	if _, err := ApplyControl(dir, ControlPatch{SessionID: "s", Goal: "ship it"}, zeroTime()); err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyControl(dir, ControlPatch{LoopTask: "repeat"}, zeroTime()); err == nil {
		t.Fatal("expected loop to be rejected while goal is active")
	}
}

func TestApplyControlStoresTopicAndYolo(t *testing.T) {
	dir := t.TempDir()
	yolo := true
	state, err := ApplyControl(dir, ControlPatch{SessionID: "s", Topic: "provider routing", Yolo: &yolo}, zeroTime())
	if err != nil {
		t.Fatal(err)
	}
	if !state.Yolo || state.Topic != "provider routing" {
		t.Fatalf("unexpected state: %#v", state)
	}
	loaded, err := LoadControl(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.Yolo || loaded.Topic != state.Topic {
		t.Fatalf("unexpected loaded state: %#v", loaded)
	}
}

func zeroTime() time.Time { return time.Time{} }
