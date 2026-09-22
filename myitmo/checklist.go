package myitmo

import (
	"context"

	"github.com/kewldan/go-itmo/internal/jsonx"
)

// ChecklistService is bypass checklist (/api/services/checklist).
type ChecklistService struct{ c *Client }

// Decision badges of ChecklistItem.DecisionID and ChecklistDepartment.DecisionID.
const (
	ChecklistDecisionProblem = 1 // debt or blocker
	ChecklistDecisionWarning = 2
	ChecklistDecisionCleared = 3
)

// ChecklistBypass is the user's clearance sheet.
type ChecklistBypass struct {
	Student     []ChecklistItem       `json:"student"`
	Departments []ChecklistDepartment `json:"departments"`
}

// ChecklistItem is a personal clearance entry.
type ChecklistItem struct {
	Problem  string `json:"problem"`
	Decision string `json:"decision"`
	// DecisionID is one of the ChecklistDecision constants.
	DecisionID int    `json:"decision_id"`
	Comment    string `json:"comment"`
	// Action is "link" (open URL) or "req" (file a service-desk request of ActionID).
	Action   string `json:"action"`
	URL      string `json:"url"`
	ActionID *int64 `json:"action_id"`
}

// ChecklistDepartment is a department's clearance entry.
type ChecklistDepartment struct {
	Department string `json:"department"`
	Problem    string `json:"problem"`
	Decision   string `json:"decision"`
	// DecisionID is one of the ChecklistDecision constants.
	DecisionID        int    `json:"decision_id"`
	ContactPersonID   int64  `json:"contact_person_id"`
	ContactPersonName string `json:"contact_person_name"`
	Comment           string `json:"comment"`
	SortOrder         int    `json:"sort_order"`
	// Actions is a list of ChecklistAction or an empty string; see ActionList.
	Actions RawJSON `json:"actions"`
}

// ChecklistAction is an action offered for a department entry.
type ChecklistAction struct {
	// Action is "req" (file a service-desk request of ActionID) or "warning".
	Action   string `json:"action"`
	ActionID int64  `json:"action_id"`
	Comment  string `json:"comment"`
}

// ActionList decodes Actions; an empty string or null yields no actions.
func (d ChecklistDepartment) ActionList() ([]ChecklistAction, error) {
	if len(d.Actions) == 0 || d.Actions[0] != '[' {
		return nil, nil
	}
	var out []ChecklistAction
	if err := jsonx.Unmarshal(d.Actions, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Bypass returns the user's clearance sheet: personal and per-department debts and required actions.
// GET /api/services/checklist/bypass
func (s *ChecklistService) Bypass(ctx context.Context) (*ChecklistBypass, error) {
	return call[*ChecklistBypass](ctx, s.c, get("api/services/checklist/bypass", nil))
}
