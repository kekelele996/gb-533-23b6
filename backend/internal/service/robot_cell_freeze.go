package service

import (
	"fmt"
	"strings"

	"robot-cell-safety-envelope-validator/backend/internal/constants"
	"robot-cell-safety-envelope-validator/backend/internal/dto"
	"robot-cell-safety-envelope-validator/backend/internal/model"
)

// Freeze issue kinds reported by the pre-freeze publish checks.
const (
	FreezeIssueNoSafetyZone           = "no_safety_zone"
	FreezeIssueZoneNotActive          = "zone_not_active"
	FreezeIssueNoReadyOrActiveProgram = "no_ready_or_active_program"
)

// BuildFreezeIssues collects the blocking reasons for freezing a cell layout:
// safety zones that are not enabled, and the absence of a ready or active
// motion program. It is a pure function so the policy can be unit tested.
func BuildFreezeIssues(zones []model.SafetyZone, programs []model.MotionProgram) []dto.FreezeCheckIssue {
	issues := []dto.FreezeCheckIssue{}
	if len(zones) == 0 {
		issues = append(issues, dto.FreezeCheckIssue{
			Kind:    FreezeIssueNoSafetyZone,
			Message: "no safety zone is defined; create and enable at least one safety zone before freezing",
		})
	}
	for _, zone := range zones {
		if zone.ZoneState == constants.ZoneStateActive {
			continue
		}
		issues = append(issues, dto.FreezeCheckIssue{
			Kind:      FreezeIssueZoneNotActive,
			Message:   fmt.Sprintf("safety zone %q is %s and not enabled; activate it or remove it before freezing", zone.Name, zoneStateLabel(zone.ZoneState)),
			ZoneID:    zone.ID,
			ZoneName:  zone.Name,
			ZoneState: zone.ZoneState,
		})
	}
	hasUsableProgram := false
	for _, program := range programs {
		if program.ProgramState == constants.ProgramStateReady || program.ProgramState == constants.ProgramStateActive {
			hasUsableProgram = true
			break
		}
	}
	if !hasUsableProgram {
		issues = append(issues, dto.FreezeCheckIssue{
			Kind:         FreezeIssueNoReadyOrActiveProgram,
			Message:      "no motion program is ready or active; import a program and advance it to ready or active before freezing",
			ProgramCount: len(programs),
		})
	}
	return issues
}

func zoneStateLabel(state string) string {
	switch state {
	case constants.ZoneStateDraft:
		return "still in draft"
	case constants.ZoneStateInactive:
		return "deactivated"
	default:
		return strings.ReplaceAll(state, "_", " ")
	}
}
