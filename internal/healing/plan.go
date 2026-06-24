package healing

import (
	"github.com/yunusemrejr/Harnejr/internal/doctor"
	"github.com/yunusemrejr/Harnejr/internal/quality"
)

type Step struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
}

type Plan struct {
	Status string `json:"status"`
	Steps  []Step `json:"steps"`
}

func BuildPlan(report doctor.Report, loc quality.LoCReport) Plan {
	plan := Plan{Status: "clear"}
	for _, check := range report.Checks {
		if check.Status != "pass" {
			plan.Steps = append(plan.Steps, Step{ID: "doctor." + check.ID, Description: check.Message, Priority: 10})
		}
	}
	for _, file := range loc.Oversized {
		plan.Steps = append(plan.Steps, Step{ID: "loc." + file.Path, Description: "Split or justify oversized source file: " + file.Path, Priority: 8})
	}
	if len(plan.Steps) > 0 {
		plan.Status = "repair_required"
	}
	return plan
}
