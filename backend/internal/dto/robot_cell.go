package dto

import (
	"encoding/json"
	"time"
)

type CreateRobotCellRequest struct {
	CellCode        string          `json:"cell_code" validate:"required,min=3,max=64"`
	Name            string          `json:"name" validate:"required,min=3,max=160"`
	LayoutGeoJSON   json.RawMessage `json:"layout_geojson" validate:"required"`
	RobotModel      string          `json:"robot_model" validate:"required,max=120"`
	ControllerModel string          `json:"controller_model" validate:"required,max=120"`
	MaxReachMM      float64         `json:"max_reach_mm" validate:"required,gt=0,lte=20000"`
	OwnerTeam       string          `json:"owner_team" validate:"required,max=120"`
}

type UpdateRobotCellRequest struct {
	Name            string          `json:"name" validate:"required,min=3,max=160"`
	LayoutGeoJSON   json.RawMessage `json:"layout_geojson" validate:"required"`
	RobotModel      string          `json:"robot_model" validate:"required,max=120"`
	ControllerModel string          `json:"controller_model" validate:"required,max=120"`
	MaxReachMM      float64         `json:"max_reach_mm" validate:"required,gt=0,lte=20000"`
	OwnerTeam       string          `json:"owner_team" validate:"required,max=120"`
	LayoutVersion   int             `json:"layout_version" validate:"required,gte=1"`
}

type RobotCellResponse struct {
	ID              uint            `json:"id"`
	CellCode        string          `json:"cell_code"`
	Name            string          `json:"name"`
	LayoutGeoJSON   json.RawMessage `json:"layout_geojson"`
	RobotModel      string          `json:"robot_model"`
	ControllerModel string          `json:"controller_model"`
	MaxReachMM      float64         `json:"max_reach_mm"`
	OwnerTeam       string          `json:"owner_team"`
	CellState       string          `json:"cell_state"`
	LayoutVersion   int             `json:"layout_version"`
	ZoneCount       int64           `json:"zone_count"`
	ProgramCount    int64           `json:"program_count"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// FreezeCheckIssue describes one blocking reason returned with a 409
// when a draft cell cannot pass the pre-freeze publish checks.
type FreezeCheckIssue struct {
	Kind         string `json:"kind"`
	Message      string `json:"message"`
	ZoneID       uint   `json:"zone_id,omitempty"`
	ZoneName     string `json:"zone_name,omitempty"`
	ZoneState    string `json:"zone_state,omitempty"`
	ProgramCount int    `json:"program_count,omitempty"`
}
