package service

import (
	"testing"

	"robot-cell-safety-envelope-validator/backend/internal/constants"
	"robot-cell-safety-envelope-validator/backend/internal/model"
)

func zone(id uint, name, state string) model.SafetyZone {
	return model.SafetyZone{ID: id, Name: name, ZoneState: state}
}

func program(id uint, code string, version int, state string) model.MotionProgram {
	return model.MotionProgram{ID: id, ProgramCode: code, Version: version, ProgramState: state}
}

func TestBuildFreezeIssues(t *testing.T) {
	tests := []struct {
		name       string
		zones      []model.SafetyZone
		programs   []model.MotionProgram
		wantKinds  []string
		wantCounts map[string]int
		wantPasses bool
	}{
		{
			name:       "active zone and active program pass",
			zones:      []model.SafetyZone{zone(1, "Gate", constants.ZoneStateActive)},
			programs:   []model.MotionProgram{program(1, "MOVE-1", 1, constants.ProgramStateActive)},
			wantPasses: true,
		},
		{
			name:       "ready program is usable",
			zones:      []model.SafetyZone{zone(1, "Gate", constants.ZoneStateActive)},
			programs:   []model.MotionProgram{program(1, "MOVE-1", 1, constants.ProgramStateReady)},
			wantPasses: true,
		},
		{
			name:       "empty cell reports both missing assets",
			zones:      nil,
			programs:   nil,
			wantKinds:  []string{FreezeIssueNoSafetyZone, FreezeIssueNoReadyOrActiveProgram},
			wantCounts: map[string]int{FreezeIssueNoSafetyZone: 1, FreezeIssueNoReadyOrActiveProgram: 1},
		},
		{
			name:       "draft and inactive zones are listed individually",
			zones:      []model.SafetyZone{zone(1, "Draft aisle", constants.ZoneStateDraft), zone(2, "Inactive gate", constants.ZoneStateInactive)},
			programs:   []model.MotionProgram{program(1, "MOVE-1", 1, constants.ProgramStateUploaded)},
			wantKinds:  []string{FreezeIssueZoneNotActive, FreezeIssueNoReadyOrActiveProgram},
			wantCounts: map[string]int{FreezeIssueZoneNotActive: 2, FreezeIssueNoReadyOrActiveProgram: 1},
		},
		{
			name:       "active zone mixed with draft zone still fails on zone",
			zones:      []model.SafetyZone{zone(1, "Gate", constants.ZoneStateActive), zone(2, "Aisle", constants.ZoneStateDraft)},
			programs:   []model.MotionProgram{program(1, "MOVE-1", 1, constants.ProgramStateReady)},
			wantKinds:  []string{FreezeIssueZoneNotActive},
			wantCounts: map[string]int{FreezeIssueZoneNotActive: 1},
		},
		{
			name:       "only uploaded or superseded programs fail the program check",
			zones:      []model.SafetyZone{zone(1, "Gate", constants.ZoneStateActive)},
			programs:   []model.MotionProgram{program(1, "OLD", 1, constants.ProgramStateSuperseded), program(2, "NEW", 2, constants.ProgramStateUploaded)},
			wantKinds:  []string{FreezeIssueNoReadyOrActiveProgram},
			wantCounts: map[string]int{FreezeIssueNoReadyOrActiveProgram: 1},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			issues := BuildFreezeIssues(test.zones, test.programs)
			if test.wantPasses {
				if len(issues) != 0 {
					t.Fatalf("expected no issues, got %d: %+v", len(issues), issues)
				}
				return
			}
			counts := map[string]int{}
			for _, issue := range issues {
				if issue.Message == "" {
					t.Fatalf("issue %q must carry a human-readable message", issue.Kind)
				}
				counts[issue.Kind]++
			}
			for _, kind := range test.wantKinds {
				if counts[kind] == 0 {
					t.Fatalf("expected issue kind %q among %+v", kind, issues)
				}
			}
			for kind, expected := range test.wantCounts {
				if counts[kind] != expected {
					t.Fatalf("expected %d issue(s) of kind %q, got %d", expected, kind, counts[kind])
				}
			}
		})
	}
}

func TestBuildFreezeIssuesDraftZoneDetails(t *testing.T) {
	issues := BuildFreezeIssues(
		[]model.SafetyZone{zone(7, "Transfer gate", constants.ZoneStateDraft)},
		nil,
	)
	var found bool
	for _, issue := range issues {
		if issue.Kind != FreezeIssueZoneNotActive {
			continue
		}
		found = true
		if issue.ZoneID != 7 || issue.ZoneName != "Transfer gate" || issue.ZoneState != constants.ZoneStateDraft {
			t.Fatalf("zone issue missing entity details: %+v", issue)
		}
	}
	if !found {
		t.Fatal("expected a zone_not_active issue")
	}
}
