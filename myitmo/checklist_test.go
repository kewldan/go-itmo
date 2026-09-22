package myitmo_test

import (
	"net/http"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestChecklistBypass(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"student":[{"problem":"Library debt","decision":"Return books","decision_id":1,"comment":"","action":"req","url":null,"action_id":7}],
		"departments":[
			{"department":"Library","problem":"","decision":"Cleared","decision_id":3,"contact_person_id":5,"contact_person_name":"Contact A","comment":"","sort_order":1,"actions":""},
			{"department":"Dormitory","problem":"Room","decision":"Check out","decision_id":2,"contact_person_id":6,"contact_person_name":"Contact B","comment":"","sort_order":2,"actions":[{"action":"req","action_id":9,"comment":"File a request"}]}]}`)
	b := must[*myitmo.ChecklistBypass](t)(c.Checklist.Bypass(ctx))
	f.expect(http.MethodGet, "/api/services/checklist/bypass")
	if len(b.Student) != 1 || b.Student[0].DecisionID != myitmo.ChecklistDecisionProblem || *b.Student[0].ActionID != 7 || b.Student[0].URL != "" {
		t.Fatalf("student = %+v", b.Student)
	}
	if len(b.Departments) != 2 || b.Departments[0].DecisionID != myitmo.ChecklistDecisionCleared {
		t.Fatalf("departments = %+v", b.Departments)
	}
	none, err := b.Departments[0].ActionList()
	if err != nil || none != nil {
		t.Errorf("empty actions = %+v, %v", none, err)
	}
	acts, err := b.Departments[1].ActionList()
	if err != nil || len(acts) != 1 || acts[0].ActionID != 9 || acts[0].Action != "req" {
		t.Errorf("actions = %+v, %v", acts, err)
	}
}
