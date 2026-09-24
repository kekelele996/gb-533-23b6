package service

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"robot-cell-safety-envelope-validator/backend/internal/constants"
	"robot-cell-safety-envelope-validator/backend/internal/dto"
	"robot-cell-safety-envelope-validator/backend/internal/model"
	"robot-cell-safety-envelope-validator/backend/internal/repository"
)

func newFreezeTestService(t *testing.T) (*RobotCellService, *gorm.DB) {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "-"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.RobotCell{}, &model.SafetyZone{}, &model.MotionProgram{}, &model.AuditEvent{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	system := NewSystemService(repository.NewSystemRepository(db), "test-secret-with-enough-length", time.Hour)
	return NewRobotCellService(repository.NewRobotCellRepository(db), system), db
}

func freezeTestActor() dto.Actor {
	return dto.Actor{ID: 1, Username: "engineer", Role: constants.RoleSafetyEngineer}
}

func createDraftCell(t *testing.T, db *gorm.DB) model.RobotCell {
	t.Helper()
	cell := model.RobotCell{
		CellCode: "CELL-T01", Name: "Freeze precheck cell", LayoutGeoJSON: `{"type":"FeatureCollection","features":[]}`,
		RobotModel: "QA-Robot", ControllerModel: "QA-Control", MaxReachMM: 2000, OwnerTeam: "QA Integration",
		CellState: constants.CellStateDraft, LayoutVersion: 1, CreatedBy: 1,
	}
	if err := db.Create(&cell).Error; err != nil {
		t.Fatalf("create cell: %v", err)
	}
	return cell
}

func requireAppError(t *testing.T, err error, status int, code string, issueCount int) []string {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %s error, got nil", code)
	}
	appError := &AppError{}
	if !errors.As(err, &appError) {
		t.Fatalf("error type %T is not AppError: %v", err, err)
	}
	if appError.Status != status || appError.Code != code {
		t.Fatalf("status=%d code=%s, want %d %s", appError.Status, appError.Code, status, code)
	}
	var issues []string
	if details, ok := appError.Details.(map[string]any); ok {
		issues, _ = details["issues"].([]string)
	}
	if len(issues) != issueCount {
		t.Fatalf("issues = %v, want %d entries", issues, issueCount)
	}
	return issues
}

func TestFreezePrecheck(t *testing.T) {
	service, db := newFreezeTestService(t)
	actor := freezeTestActor()
	cell := createDraftCell(t, db)

	issues := requireAppError(t, mustFreezeErr(service, cell.ID, actor, "req-empty"), http.StatusConflict, "freeze_precheck_failed", 2)
	if !strings.Contains(issues[0], "no safety zones") || !strings.Contains(issues[1], "no motion program") {
		t.Fatalf("unexpected issues for empty cell: %v", issues)
	}
	reloaded, err := service.repository.Get(cell.ID)
	if err != nil || reloaded.CellState != constants.CellStateDraft {
		t.Fatalf("failed precheck must not change state: state=%s err=%v", reloaded.CellState, err)
	}

	zone := model.SafetyZone{
		RobotCellID: cell.ID, Name: "North gate", ZoneType: constants.ZoneTypeRestricted,
		PolygonGeoJSON: `{"type":"Polygon","coordinates":[[[0,0],[10,0],[10,10],[0,0]]]}`,
		MinHeightMM:    0, MaxHeightMM: 2000, SpeedLimitMMS: 100, AccessRule: "Gate lock must precede motion",
		ZoneState: constants.ZoneStateDraft, Version: 1, CreatedBy: 1,
	}
	if err := db.Create(&zone).Error; err != nil {
		t.Fatalf("create zone: %v", err)
	}
	issues = requireAppError(t, mustFreezeErr(service, cell.ID, actor, "req-draft-zone"), http.StatusConflict, "freeze_precheck_failed", 2)
	zoneListed := false
	for _, issue := range issues {
		zoneListed = zoneListed || strings.Contains(issue, "North gate")
	}
	if !zoneListed {
		t.Fatalf("inactive zone must be listed by name: %v", issues)
	}

	if err := db.Model(&model.SafetyZone{}).Where("id = ?", zone.ID).Update("zone_state", constants.ZoneStateActive).Error; err != nil {
		t.Fatalf("activate zone: %v", err)
	}
	program := model.MotionProgram{
		RobotCellID: cell.ID, ProgramCode: "QA-MOVE-T01", Version: 1, TrajectoryJSON: "[]",
		InterlockSequenceJSON: "[]", SourceChecksum: "checksum", ProgramState: constants.ProgramStateParsed,
		UploadedBy: 1, UploadedAt: time.Now().UTC(),
	}
	if err := db.Create(&program).Error; err != nil {
		t.Fatalf("create program: %v", err)
	}
	requireAppError(t, mustFreezeErr(service, cell.ID, actor, "req-parsed-program"), http.StatusConflict, "freeze_precheck_failed", 1)

	if err := db.Model(&model.MotionProgram{}).Where("id = ?", program.ID).Update("program_state", constants.ProgramStateReady).Error; err != nil {
		t.Fatalf("ready program: %v", err)
	}
	response, err := service.Freeze(cell.ID, actor, "req-pass")
	if err != nil {
		t.Fatalf("freeze should pass once issues are resolved: %v", err)
	}
	if response.CellState != constants.CellStateFrozen {
		t.Fatalf("cell_state = %s, want frozen", response.CellState)
	}

	requireAppError(t, mustFreezeErr(service, cell.ID, actor, "req-repeat"), http.StatusConflict, "state_conflict", 0)
}

func mustFreezeErr(service *RobotCellService, id uint, actor dto.Actor, requestID string) error {
	_, err := service.Freeze(id, actor, requestID)
	return err
}
